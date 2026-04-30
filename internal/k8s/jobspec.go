package k8s

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rustyguts/tidal/internal/domain"
)

// JobSpec carries everything needed to template a `batch/v1` Job.
type JobSpec struct {
	JobID             domain.JobID
	Namespace         string
	Image             string
	ImagePullPolicy   corev1.PullPolicy
	ServiceAccount    string
	ServerInternalURL string
	// MediaPVC names a PersistentVolumeClaim that the per-job pod should mount
	// at MediaMountPath. The chart binds the claim to whatever storage class
	// the operator chose; the dispatcher only sees the claim name.
	MediaPVC         string
	MediaHostPath    string // optional dev fallback when no PVC is available
	MediaMountPath   string // default /media
	BackoffLimit     int32
	TTLSecondsAfter  int32
	Resources        corev1.ResourceRequirements
	NodeSelector     map[string]string
	Tolerations      []corev1.Toleration
	RuntimeClassName string
}

// Build returns a Kubernetes Job object. Caller submits via batch/v1
// Jobs(namespace).Create.
func Build(spec JobSpec) *batchv1.Job {
	mountPath := spec.MediaMountPath
	if mountPath == "" {
		mountPath = "/media"
	}
	pull := spec.ImagePullPolicy
	if pull == "" {
		pull = corev1.PullIfNotPresent
	}

	volumes, mounts := buildMediaVolume(spec, mountPath)

	env := []corev1.EnvVar{
		{Name: "TIDAL_LOG_PRETTY", Value: "false"},
	}

	backoff := spec.BackoffLimit
	ttl := spec.TTLSecondsAfter
	if ttl == 0 {
		ttl = 600
	}

	args := []string{
		"runjob",
		"--job-id=" + spec.JobID.String(),
		"--server-url=" + spec.ServerInternalURL,
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tidal-job-" + spec.JobID.String(),
			Namespace: spec.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "tidal",
				"app.kubernetes.io/component":  "transcode-job",
				"app.kubernetes.io/managed-by": "tidal-dispatcher",
				"tidal.io/job-id":              spec.JobID.String(),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoff,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      "tidal",
						"app.kubernetes.io/component": "transcode-job",
						"tidal.io/job-id":             spec.JobID.String(),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: spec.ServiceAccount,
					NodeSelector:       spec.NodeSelector,
					Tolerations:        spec.Tolerations,
					Volumes:            volumes,
					Containers: []corev1.Container{{
						Name:            "tidal",
						Image:           spec.Image,
						ImagePullPolicy: pull,
						Command:         []string{"tidal"},
						Args:            args,
						Env:             env,
						VolumeMounts:    mounts,
						Resources:       spec.Resources,
					}},
				},
			},
		},
	}
	if spec.RuntimeClassName != "" {
		rc := spec.RuntimeClassName
		job.Spec.Template.Spec.RuntimeClassName = &rc
	}
	return job
}

func buildMediaVolume(spec JobSpec, mountPath string) ([]corev1.Volume, []corev1.VolumeMount) {
	mount := []corev1.VolumeMount{{Name: "media", MountPath: mountPath}}
	if spec.MediaPVC != "" {
		return []corev1.Volume{{
			Name: "media",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: spec.MediaPVC,
				},
			},
		}}, mount
	}
	if spec.MediaHostPath != "" {
		hp := corev1.HostPathDirectoryOrCreate
		return []corev1.Volume{{
			Name: "media",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{Path: spec.MediaHostPath, Type: &hp},
			},
		}}, mount
	}
	return nil, nil
}
