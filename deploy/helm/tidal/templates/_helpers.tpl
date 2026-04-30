{{/*
Common helpers
*/}}

{{- define "tidal.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "tidal.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tidal.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "tidal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "tidal.serverSelectorLabels" -}}
app.kubernetes.io/name: {{ include "tidal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: server
{{- end -}}

{{- define "tidal.dispatcherSelectorLabels" -}}
app.kubernetes.io/name: {{ include "tidal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: dispatcher
{{- end -}}

{{- define "tidal.workerSelectorLabels" -}}
app.kubernetes.io/name: {{ include "tidal.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: worker
{{- end -}}

{{- define "tidal.image" -}}
{{- $tag := default .Chart.AppVersion .Values.image.tag -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}

{{- define "tidal.serverServiceAccount" -}}
{{- if .Values.serviceAccount.server.create -}}
{{- default (printf "%s-server" (include "tidal.fullname" .)) .Values.serviceAccount.server.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.server.name -}}
{{- end -}}
{{- end -}}

{{- define "tidal.dispatcherServiceAccount" -}}
{{- if .Values.serviceAccount.dispatcher.create -}}
{{- default (printf "%s-dispatcher" (include "tidal.fullname" .)) .Values.serviceAccount.dispatcher.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.dispatcher.name -}}
{{- end -}}
{{- end -}}

{{- define "tidal.workerServiceAccount" -}}
{{- if .Values.serviceAccount.worker.create -}}
{{- default (printf "%s-worker" (include "tidal.fullname" .)) .Values.serviceAccount.worker.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.worker.name -}}
{{- end -}}
{{- end -}}

{{- define "tidal.callbackSecretName" -}}
{{- if .Values.callbackSecret.existingSecret -}}
{{- .Values.callbackSecret.existingSecret -}}
{{- else -}}
{{- printf "%s-callback" (include "tidal.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "tidal.databaseSecretName" -}}
{{- if .Values.externalDatabase.existingSecret -}}
{{- .Values.externalDatabase.existingSecret -}}
{{- else -}}
{{- printf "%s-db" (include "tidal.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "tidal.redisSecretName" -}}
{{- if .Values.externalRedis.existingSecret -}}
{{- .Values.externalRedis.existingSecret -}}
{{- else -}}
{{- printf "%s-redis" (include "tidal.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/* Common env vars wired into both server + dispatcher pods. */}}
{{- define "tidal.commonEnv" -}}
- name: TIDAL_LOG_PRETTY
  value: "false"
- name: TIDAL_DB_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "tidal.databaseSecretName" . }}
      key: database-url
- name: TIDAL_REDIS_URL
  valueFrom:
    secretKeyRef:
      name: {{ include "tidal.redisSecretName" . }}
      key: redis-url
- name: TIDAL_CALLBACK_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "tidal.callbackSecretName" . }}
      key: callback-secret
{{- end -}}

{{/* tidal.mediaPVC returns the effective claim name the chart hands to Tidal.
     Empty string means "use hostPath fallback". */}}
{{- define "tidal.mediaPVC" -}}
{{- if .Values.media.pvc.enabled -}}
{{- printf "%s-media" (include "tidal.fullname" .) -}}
{{- else if .Values.media.existingClaim -}}
{{- .Values.media.existingClaim -}}
{{- else if and .Values.media.nfs.server .Values.media.nfs.path -}}
{{- printf "%s-media-nfs" (include "tidal.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "tidal.dispatcherEnv" -}}
{{- if eq .Values.dispatcher.mode "batch" }}
- name: TIDAL_DISPATCHER
  value: k8s
- name: TIDAL_DISPATCHER_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: TIDAL_DISPATCHER_JOB_IMAGE
  value: {{ default (include "tidal.image" .) .Values.dispatcher.jobTemplate.image | quote }}
- name: TIDAL_DISPATCHER_JOB_SERVICE_ACCOUNT
  value: {{ include "tidal.dispatcherServiceAccount" . }}
- name: TIDAL_SERVER_INTERNAL_URL
  value: http://{{ include "tidal.fullname" . }}-server.{{ .Release.Namespace }}.svc.cluster.local:{{ .Values.service.port }}
{{- $pvc := include "tidal.mediaPVC" . }}
{{- if $pvc }}
- name: TIDAL_DISPATCHER_MEDIA_PVC
  value: {{ $pvc | quote }}
{{- else if .Values.media.hostPath }}
- name: TIDAL_DISPATCHER_MEDIA_HOSTPATH
  value: {{ .Values.media.hostPath | quote }}
{{- end }}
- name: TIDAL_DISPATCHER_MEDIA_MOUNT_PATH
  value: {{ .Values.media.mountPath | quote }}
{{- with .Values.dispatcher.jobTemplate.resources.requests }}
{{- with .cpu }}
- name: TIDAL_DISPATCHER_JOB_REQUEST_CPU
  value: {{ . | quote }}
{{- end }}
{{- with .memory }}
- name: TIDAL_DISPATCHER_JOB_REQUEST_MEMORY
  value: {{ . | quote }}
{{- end }}
{{- end }}
{{- with .Values.dispatcher.jobTemplate.resources.limits }}
{{- with .cpu }}
- name: TIDAL_DISPATCHER_JOB_LIMIT_CPU
  value: {{ . | quote }}
{{- end }}
{{- with .memory }}
- name: TIDAL_DISPATCHER_JOB_LIMIT_MEMORY
  value: {{ . | quote }}
{{- end }}
{{- end }}
{{- end }}
- name: TIDAL_WORKER_CONCURRENCY
  value: {{ .Values.dispatcher.workerConcurrency | quote }}
{{- end -}}

{{- define "tidal.workerEnv" -}}
- name: TIDAL_DISPATCHER
  value: local
- name: TIDAL_WORKER_CONCURRENCY
  value: {{ .Values.worker.concurrency | quote }}
{{- end -}}

{{- define "tidal.mediaVolumeMounts" -}}
{{- $pvc := include "tidal.mediaPVC" . -}}
{{- if or $pvc .Values.media.hostPath }}
- name: media
  mountPath: {{ .Values.media.mountPath | quote }}
{{- end -}}
{{- end -}}

{{- define "tidal.mediaVolumes" -}}
{{- $pvc := include "tidal.mediaPVC" . -}}
{{- if $pvc }}
- name: media
  persistentVolumeClaim:
    claimName: {{ $pvc | quote }}
{{- else if .Values.media.hostPath }}
- name: media
  hostPath:
    path: {{ .Values.media.hostPath | quote }}
    type: DirectoryOrCreate
{{- end -}}
{{- end -}}
