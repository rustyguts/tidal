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

// Dispatcher creates one batch/v1.Job per task and waits for it to land.
// Per-job progress streams over REST callbacks (handled by `tidal runjob`),
// not back through this code.
type Dispatcher struct {
	cli       *kubernetes.Clientset
	specProto JobSpec
}

func NewDispatcher(cli *kubernetes.Clientset, proto JobSpec) *Dispatcher {
	return &Dispatcher{cli: cli, specProto: proto}
}

// Run creates the K8s Job for the given JobID and blocks until the Job lands
// (Succeeded or Failed). Returns nil on success, error on failure or context
// cancellation. The asynq handler should propagate ctx so cancellation
// triggers Job deletion.
func (d *Dispatcher) Run(ctx context.Context, jobID domain.JobID) error {
	spec := d.specProto
	spec.JobID = jobID

	job := Build(spec)
	created, err := d.cli.BatchV1().Jobs(spec.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Idempotent re-dispatch (asynq retried the task while a previous Job
			// is still around). Fetch the existing one and watch it.
			created, err = d.cli.BatchV1().Jobs(spec.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("get existing job: %w", err)
			}
		} else {
			return fmt.Errorf("create job: %w", err)
		}
	}
	log.Info().Str("job", string(jobID.String())).Str("k8sJob", created.Name).Msg("k8s job created")

	defer func() {
		// best-effort cleanup on cancel; TTLSecondsAfterFinished handles success.
		if errors.Is(ctx.Err(), context.Canceled) {
			cctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			background := metav1.DeletePropagationBackground
			err := d.cli.BatchV1().Jobs(spec.Namespace).Delete(cctx, created.Name, metav1.DeleteOptions{
				PropagationPolicy: &background,
			})
			if err != nil && !apierrors.IsNotFound(err) {
				log.Warn().Err(err).Str("k8sJob", created.Name).Msg("delete cancelled job")
			}
		}
	}()

	return d.waitForCompletion(ctx, spec.Namespace, created.Name)
}

// waitForCompletion polls the Job phase until terminal. A real implementation
// should use an Informer; polling keeps the dependency surface small.
func (d *Dispatcher) waitForCompletion(ctx context.Context, namespace, name string) error {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tick.C:
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
