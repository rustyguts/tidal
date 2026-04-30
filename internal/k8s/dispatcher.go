package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/rs/zerolog/log"

	"github.com/rustyguts/tidal/internal/domain"
)

// JobCoordinator is the dispatcher's hook back into the Tidal jobs.Service.
// The dispatcher uses it to:
//   - poll for user-initiated cancel intent (status=cancelling/cancelled),
//   - mark the DB row terminal once it observes a K8s Job condition or
//     completes a user-cancel-driven delete.
//
// runjob no longer reports terminal state on pod SIGTERM (pod termination !=
// user cancel), so the dispatcher is the authoritative writer for terminal
// status in the k8s dispatcher mode.
type JobCoordinator interface {
	IsCancelRequested(ctx context.Context, jobID domain.JobID) (bool, error)
	Cancelled(ctx context.Context, jobID domain.JobID)
	Failed(ctx context.Context, jobID domain.JobID, jobErr error)
}

// Dispatcher creates one batch/v1.Job per task and watches it to completion.
// Per-job progress streams over REST callbacks (handled by `tidal runjob`).
//
// Process restarts (SIGTERM) do NOT delete in-flight K8s Jobs — they are
// allowed to keep running. asynq retries the task; the new dispatcher pod
// re-attaches via the deterministic Job name and resumes watching.
type Dispatcher struct {
	cli         *kubernetes.Clientset
	specProto   JobSpec
	coordinator JobCoordinator
}

func NewDispatcher(cli *kubernetes.Clientset, proto JobSpec) *Dispatcher {
	return &Dispatcher{cli: cli, specProto: proto}
}

// SetCoordinator wires the bridge to jobs.Service for cancel intent + DB
// terminal-status writes. Required for k8s dispatcher mode; if nil the
// dispatcher cannot finalize jobs and the DB row stays "running" forever.
func (d *Dispatcher) SetCoordinator(c JobCoordinator) { d.coordinator = c }

// Run creates the K8s Job for the given JobID and blocks until terminal.
// All DB terminal-state writes (cancelled/failed/succeeded) flow through the
// coordinator — runjob does not write terminal state in k8s mode.
//
// Returns:
//   - nil on K8s Job Complete (runjob already posted Succeeded)
//   - nil on K8s Job Failed (coordinator writes Failed; no asynq retry)
//   - nil on observed user cancel (coordinator writes Cancelled)
//   - ctx.Err() on dispatcher shutdown (Job left running; asynq retries re-attach)
func (d *Dispatcher) Run(ctx context.Context, jobID domain.JobID) error {
	spec := d.specProto
	spec.JobID = jobID

	job := Build(spec)
	created, err := d.cli.BatchV1().Jobs(spec.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Idempotent re-dispatch — asynq retried us after a dispatcher
			// restart, or a sibling already created the Job. Re-attach.
			created, err = d.cli.BatchV1().Jobs(spec.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get existing job: %w", err)
			}
		} else {
			return fmt.Errorf("create job: %w", err)
		}
	}
	log.Info().Str("job", jobID.String()).Str("k8sJob", created.Name).Msg("k8s job watching")

	return d.waitForCompletion(ctx, spec.Namespace, created.Name, jobID)
}

// waitForCompletion polls the Job phase + the DB cancel-request flag.
// Returns ctx.Err() on dispatcher shutdown WITHOUT touching the K8s Job —
// the in-flight encode keeps running across rolling restarts; asynq's retry
// path will re-attach.
//
// On any observed terminal outcome (user cancel, K8s Job failed, K8s Job
// disappeared) the coordinator writes the DB row to its terminal state and
// the dispatcher returns nil so asynq does NOT retry the task.
func (d *Dispatcher) waitForCompletion(ctx context.Context, namespace, name string, jobID domain.JobID) error {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			// User-initiated cancel.
			if d.coordinator != nil {
				cancelReq, err := d.coordinator.IsCancelRequested(ctx, jobID)
				if err == nil && cancelReq {
					d.deleteJob(namespace, name)
					d.coordinator.Cancelled(context.Background(), jobID)
					return nil
				}
			}

			job, err := d.cli.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					// Job vanished without a terminal condition — most likely
					// pruned by an external controller (ArgoCD prune, manual
					// kubectl delete). Treat as a hard failure rather than a
					// retry-able error so the user sees a definitive state.
					d.markFailed(jobID, errors.New("k8s Job disappeared"))
					return nil
				}
				log.Warn().Err(err).Str("k8sJob", name).Msg("get job poll")
				continue
			}
			if cond, ok := terminalCondition(job); ok {
				if cond.Type == batchv1.JobComplete {
					return nil
				}
				d.markFailed(jobID, fmt.Errorf("k8s Job failed: %s", cond.Message))
				return nil
			}
		}
	}
}

func (d *Dispatcher) markFailed(jobID domain.JobID, jobErr error) {
	log.Warn().Err(jobErr).Str("job", jobID.String()).Msg("k8s dispatch failed")
	if d.coordinator != nil {
		d.coordinator.Failed(context.Background(), jobID, jobErr)
	}
}

// deleteJob removes the K8s Job using a fresh background context — the calling
// ctx may already be cancelled. Pod is reaped via PropagationPolicy.
func (d *Dispatcher) deleteJob(namespace, name string) {
	cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	background := metav1.DeletePropagationBackground
	err := d.cli.BatchV1().Jobs(namespace).Delete(cctx, name, metav1.DeleteOptions{
		PropagationPolicy: &background,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		log.Warn().Err(err).Str("k8sJob", name).Msg("delete k8s job")
	}
}

func terminalCondition(job *batchv1.Job) (batchv1.JobCondition, bool) {
	for _, c := range job.Status.Conditions {
		if c.Status != "True" {
			continue
		}
		if c.Type == batchv1.JobComplete || c.Type == batchv1.JobFailed {
			return c, true
		}
	}
	return batchv1.JobCondition{}, false
}
