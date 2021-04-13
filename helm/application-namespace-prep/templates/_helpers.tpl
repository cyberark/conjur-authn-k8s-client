{{/* vim: set filetype=mustache: */}}
{{/*

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "conjur-prep.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Return the most recent RBAC API available
*/}}
{{- define "conjur-prep.rbac-api" -}}
{{- if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1" }}
{{- printf "rbac.authorization.k8s.io/v1" -}}
{{- else if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1alpha1" }}
{{- printf "rbac.authorization.k8s.io/v1alpha1" -}}
{{- else }}
{{- printf "rbac.authorization.k8s.io/v1" -}}
{{- end }}
{{- end }}
