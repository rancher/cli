package templates

const (
	TLSModeRancher      = "rancher"
	TLSModeIngress      = "ingress"
	TLSModeExternal     = "external"
	TLSManager          = "cert-manager"
	TLSIssuerAcme       = "acme"
	TLSIssuerCA         = "ca"
	TLSIssuerKind       = "Issuer"
	TLSIssuerSecret     = "cattle-issuer-ca"
	TLSIssuerName       = "cattle-issuer"
	TLSIngressSecret    = "default-ingress-keys"
	CattleIngressSecret = "cattle-ingress-keys"
	CattleServerSecret  = "cattle-server-keys"

	CattleHANamespace = `---
kind: Namespace
apiVersion: v1
metadata:
  name: {{.Namespace}}`

	CattleHARBAC = `---
kind: ServiceAccount
apiVersion: v1
metadata:
  name: cattle-admin
  namespace: {{.Namespace}}
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cattle-crb
  namespace: {{.Namespace}}
subjects:
- kind: ServiceAccount
  name: cattle-admin
  namespace: {{.Namespace}}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io`

	CattleHAService = `---
apiVersion: v1
kind: Service
metadata:
  namespace: {{.Namespace}}
  name: cattle-service
  labels:
    app: cattle
spec:
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
    name: http
  - port: 443
    targetPort: 443
    protocol: TCP
    name: https
  selector:
    app: cattle`

	CattleHASecret = `{{- if eq .TLS.Mode "` + TLSModeExternal + `"}}
  {{- if ne .TLS.Issuer.CACert ""}}
---
apiVersion: v1
kind: Secret
metadata:
  name: ` + CattleServerSecret + `
  namespace: {{.Namespace}}
type: Opaque
data:
  cacerts.pem: {{.TLS.Issuer.CACert}}
  {{- end}}
{{- end}}
{{- if eq .TLS.Mode "` + TLSModeIngress + `"}}
---
apiVersion: v1
kind: Secret
metadata:
  name: ` + CattleServerSecret + `
  namespace: {{.Namespace}}
type: Opaque
data:
  {{- if ne .TLS.Issuer.CACert ""}}
  cacerts.pem: {{.TLS.Issuer.CACert}}
  {{- end}}
  {{- if eq .TLS.Issuer.Kind ""}}
  cert.pem: {{.TLS.Cert}}
  key.pem: {{.TLS.Key}}
---
apiVersion: v1
kind: Secret
metadata:
  name: ` + CattleIngressSecret + `
  namespace: {{.Namespace}}
type: Opaque
data: 
  tls.crt: {{.TLS.Cert}}
  tls.key: {{.TLS.Key}}
  {{- end}}
` + CattleHAIssuer + `
{{- end}}`

	CattleHAIssuer = `{{- if ne .TLS.Issuer.Kind ""}}
  {{- if eq .TLS.Issuer.Kind "` + TLSIssuerCA + `"}}
---
apiVersion: v1
kind: Secret
metadata:
  name: ` + TLSIssuerSecret + `
  namespace: {{.Namespace}}
type: Opaque
data:
  tls.crt: {{.TLS.Issuer.CACert}}
  tls.key: {{.TLS.Issuer.CAKey}}
  {{- end}}
  {{- if .TLS.Issuer.Default}}
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: default-cert
  namespace: {{.Namespace}}
spec:
  secretName: ` + TLSIngressSecret + `
  issuerRef:
    name: ` + TLSIssuerName + `
    kind: ` + TLSIssuerKind + `
  commonName: {{.FQDN}}
  dnsNames:
  - {{.FQDN}}
  {{- end}}
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: ` + TLSIssuerKind + `
metadata:
  name: ` + TLSIssuerName + `
  namespace: {{.Namespace}}
spec:
  {{- if eq .TLS.Issuer.Kind "` + TLSIssuerCA + `"}}
  ca:
    secretName: ` + TLSIssuerSecret + `
  {{- end}}
  {{- if eq .TLS.Issuer.Kind "` + TLSIssuerAcme + `"}}
  acme:
    {{- if .TLS.Issuer.Production }}
    server: https://acme-v02.api.letsencrypt.org/directory
    {{- else }}
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    {{- end }}
    email: {{.TLS.Issuer.Email}}
    privateKeySecretRef:
      name: cattle-issuer-acme
    http01: {}
  {{- end}}
---
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: cattle-cert
  namespace: {{.Namespace}}
spec:
  secretName: ` + CattleIngressSecret + `
  issuerRef:
    name: ` + TLSIssuerName + `
    kind: ` + TLSIssuerKind + `
  commonName: {{.FQDN}}
  dnsNames:
  - {{.FQDN}}
  {{- if eq .TLS.Issuer.Kind "` + TLSIssuerAcme + `"}}
  acme:
    config:
    - http01:
        ingressClass: nginx
      domains:
      - {{.FQDN}}
  {{- end}}
{{- end}}`

	CattleHAIngress = `---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: {{.Namespace}}
  name: cattle-ingress-http
  annotations:
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "30"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "1800"   # Max time in seconds for ws to remain shell window open
    nginx.ingress.kubernetes.io/proxy-send-timeout: "1800"   # Max time in seconds for ws to remain shell window open
  {{- if eq .TLS.Mode "` + TLSModeExternal + `"}}
    nginx.ingress.kubernetes.io/ssl-redirect: "false"        # Disable redirect to ssl
  {{- end}}
  {{- if eq .TLS.Mode "` + TLSModeRancher + `"}}
    nginx.ingress.kubernetes.io/ssl-passthrough: "true"      # Enable ssl-passthrough to backend.
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"   # Force redirect to ssl.
  {{- end}}
spec:
  rules:
  - host: {{.FQDN}}
    http:
      paths:
      - backend:
          serviceName: cattle-service
        {{- if or (eq .TLS.Mode "` + TLSModeExternal + `") (eq .TLS.Mode "` + TLSModeIngress + `") }}
          servicePort: 80
        {{- end}}
        {{- if eq .TLS.Mode "` + TLSModeRancher + `"}}
          servicePort: 443
        {{- end}}
{{- if or (eq .TLS.Mode "` + TLSModeIngress + `") }}
  tls:
  - secretName: ` + CattleIngressSecret + `
    hosts: 
    - {{.FQDN}}
{{- end}}`

	CattleHADeployment = `---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  namespace: {{.Namespace}}
  name: cattle
spec:
  replicas: {{.Replicas}}
  template:
    metadata:
      labels:
        app: cattle
    spec:
      serviceAccountName: cattle-admin
      containers:
      - image: {{.Image}}
        imagePullPolicy: Always
        name: cattle-server
      {{- if and (eq .TLS.Issuer.CACert "") (ne .TLS.Mode "` + TLSModeRancher + `")}} 
        args: 
        - --no-cacerts
      {{- end}}
        env:
        - name: CATTLE_SERVER_URL
          value: https://{{.FQDN}}
        ports:
        - containerPort: 80
          protocol: TCP
        - containerPort: 443
          protocol: TCP
        livenessProbe:
          tcpSocket:
            port: 80
          initialDelaySeconds: 60
          periodSeconds: 30
        readinessProbe:
          tcpSocket:
            port: 80
          initialDelaySeconds: 20
          periodSeconds: 30
      {{- if or (ne .TLS.Issuer.CACert "") (eq .TLS.Mode "` + TLSModeIngress + `")}}
        volumeMounts:
        - mountPath: /etc/rancher/ssl
          name: ` + CattleServerSecret + `-volume
          readOnly: true
      volumes:
      - name: ` + CattleServerSecret + `-volume
        secret:
          defaultMode: 420
          secretName: ` + CattleServerSecret + `
      {{- end}}`

	CattleHATemplate = `# Cattle HA definition for {{.Mode}} tls-mode {{.TLS.Mode}}
` + CattleHANamespace + `
` + CattleHARBAC + `
` + CattleHASecret + `
` + CattleHADeployment + `
` + CattleHAService + `
` + CattleHAIngress
)
