# cert-manager-webhook-ipv64

Cert-Manager Webhook for DynDNS Provider ipv64.net

## Disclamer

This code is probably not written well since Go is a new language for me.
It works on my machine.
Please don't expect regular updates or improvements,
but feel free to create MRs/PRs or fork entirely.

The project is build around a privately hosted Gitlab instance,
hence there is no Github CI besides container image creation.

## Limitations

Since the ipv64.net do not allow more than 4 (level4.level3.level2.tld) levels of domains,
certificates can only get created for one entry on level 4 per entry of level 3,
or you use a wildcard on level 4.

## Usage

Since using ArgoCD instead of plain Helm, you might have to adapt.

### Installation of Webhook

```yaml
---
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: cert-manager-webhook-ipv64
  namespace: argocd
spec:
  destination:
    name: ''
    namespace: cert-manager
    server: 'https://kubernetes.default.svc'
  source:
    path: ''
    repoURL: 'http://chartmuseum-svc.chartmuseum.svc.cluster.local:8080' # has to match you Helm repo
    targetRevision: 1.80.0
    chart: cert-manager-webhook-ipv64
    helm:
      releaseName: cert-manager
      valuesObject:
        groupName: cert-manager-webhook-ipv64 # has to match the groupname from Issuer
        fullnameOverride: cert-manager-webhook-ipv64
        nameOverride: webhook-ipv64
        image:
          pullPolicy: Always
  sources: []
  project: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - ServerSideApply=true
    retry:
      limit: 2
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m0s
```

AI translates it to following command, but not yet testet:

`helm install cert-manager-certmanager-webhook-ipv64 <path-to-repo>/cert-manager-webhook-ipv64/Chart.yaml --values cert-manager-webhook-ipv64/values.yaml`

### Cluster Issuer

```yaml
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: acme-issuer
  namespace: cert-manager
spec:
  acme:
    privateKeySecretRef:
      name: acme-private-key
    server: https://acme-v02.api.letsencrypt.org/directory
    solvers:
    - dns01:
        webhook:
          groupName: cert-manager-webhook-ipv64 # has to match the groupname from values file
          solverName: cert-manager-webhook-ipv64 # this is a fixed value
          config:
            secretName: ipv64-token # has to match your secret name
            subdomain: my.ipv64-subdomain.de
            email: my@email.de
```

### Secret

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: ipv64-token
  namespace: cert-manager
type: Opaque
data:
  api-key: "<base64 encoded secret>" # the key has to be 'api-key'
```

## ToDos and Ideas

- [ ] define secret key via values file
- [ ] somehow deal with the rate limit usage
- [ ] improve testing / linting / security scanning
- [ ] Github Helm Repo
- [ ] Fix description of standard helm install