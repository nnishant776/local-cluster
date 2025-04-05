{{- define "dnsserver.helm_annotations" }}
meta.helm.sh/release-name: {{ .Release.Name }}
meta.helm.sh/release-namespace: {{ .Release.Namespace }}
{{- end }}

{{- define "dnsserver.helm_labels" }}
app.kubernetes.io/managed-by: Helm
{{- end }}
