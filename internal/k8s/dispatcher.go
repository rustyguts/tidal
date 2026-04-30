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

// CancelChecker reports whether the user has requested cancellation for a job
// (status == cancelling/cancelled in the DB). The dispatcher polls this from
// its watch loop and, on observed cancel intent, deletes the K8s Job.
type CancelChecker interface {
	IsCancelRequested(ctx context.Context, jobID domain.JobID) (bool, error)
}

// Dispatcher creates one batch/v1.Job per task and watches it to completion.
// Per-job progress streams over REST callbacks (handled by `tidal runjob`).
//
// Process restarts (SIGTERM) do NOT delete in-flight K8s Jobs — they are
// allowed to keep running. asynq retries the task; the new dispatcher pod
// re-attaches via the deterministic Job name and resumes watching.
type Dispatcher struct {
	cli           *kubernetes.Clientset
	specProto     JobSpec
	cancelChecker CancelChecker
}

func NewDispatcher(cli *kubernetes.Clientset, proto JobSpec) *Dispatcher {
	return &Dispatcher{cli: cli, specProto: proto}
}

// SetCancelChecker wires DB-status polling for user-initiated cancellations.
// Optional: if nil, the dispatcher only stops watching on K8s Job terminal
// state or context cancellation.
func (d *Dispatcher) SetCancelChecker(c CancelChecker) { d.cancelChecker = c }

// Run creates the K8s Job for the given JobID and blocks until terminal.
// Returns:
//   - nil on K8s Job Complete
//   - error on K8s Job Failed
//   - context.Canceled on dispatcher shutdown (Job left running for retry)
//   - "user cancelled" when CancelChecker observes status=cancelling
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
func (d *Dispatcher) waitForCompletion(ctx context.Context, namespace, name string, jobID domain.JobID) error {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
			// User-initiated cancel.
			if d.cancelChecker != nil {
				cancelReq, err := d.cancelChecker.IsCancelRequested(ctx, jobID)
				if err == nil && cancelReq {
					d.deleteJob(namespace, name)
					return errors.New("user cancelled")
				}
			}

			job, err := d.cli.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return errors.New("k8s Job disappeared")
				}
				log.Warn().Err(err).Str("k8sJob", name).Msg("get job poll")
				continue
			}
			if cond, ok := terminalCondition(job); ok {
				if cond.Type == batchv1.JobComplete {
					return nil
				}
				return fmt.Errorf("k8s Job failed: %s", cond.Message)
			}
		}
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
