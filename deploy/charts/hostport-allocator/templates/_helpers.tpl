{{/*
Expand the name of the chart.
*/}}
{{- define "hostport-allocator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hostport-allocator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "hostport-allocator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hostport-allocator.labels" -}}
helm.sh/chart: {{ include "hostport-allocator.chart" . }}
{{ include "hostport-allocator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hostport-allocator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hostport-allocator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "hostport-allocator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hostport-allocator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the certificate to use
*/}}
{{- define "hostport-allocator.certificateName" -}}
{{- default (include "hostport-allocator.fullname" .) .Values.webhook.certificate.name }}
{{- end }}

{{/*
Create the name of the certificate to use
*/}}
{{- define "hostport-allocator.certificateSecretName" -}}
{{- default (include "hostport-allocator.fullname" .) .Values.webhook.certificate.secretName }}
{{- end }}
