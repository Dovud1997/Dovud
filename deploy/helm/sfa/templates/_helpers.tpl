{{- define "sfa.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "sfa.fullname" -}}
{{- printf "%s" (include "sfa.name" .) -}}
{{- end -}}

{{- define "sfa.labels" -}}
app.kubernetes.io/name: {{ include "sfa.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
