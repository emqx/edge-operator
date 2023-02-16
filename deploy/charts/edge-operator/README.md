# edge-operator

The EDGE Operator provides [Kubernetes](https://kubernetes.io/) native deployment and management of [edge]. The purpose of this project is to simplify and automate the configuration of the edge.

## Prerequisites

- Kubernetes 1.20+

## Installing the Chart

To install the chart through github repository

```console
## clone repository
$ git clone https://github.com/emqx/edge-operator
$ cd deploy/charts/edge-operator


## Install the edge-operator helm chart
$ helm install edge-operator . \
      --namespace edge-operator-system \
      --create-namespace
```

> **Tip**: List all releases using `helm ls -A`

## Uninstalling the Chart

To uninstall/delete the `edge-operator` deployment:

```console
$ helm delete edge-operator -n edge-operator-system
```

## Configuration

The following table lists the configurable parameters of the cert-manager chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `image.repository` | Image repository | `edge/edge-operator-controller` |
| `image.tag` | Image tag | `{{RELEASE_VERSION}}` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets| `[]` |
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Default fully qualified app name. | `""` |
| `replicaCount`  | Number of cert-manager replicas  | `1` |
| `serviceAccount.create` | If `true`, create a new service account | `true` |
| `serviceAccount.name` | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the fullname template |  |
| `serviceAccount.annotations` | Annotations to add to the service account |  |
| `resources` | CPU/memory resource requests/limits | `{}` |
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `affinity` | Node affinity for pod assignment | `{}` |
| `tolerations` | Node tolerations for pod assignment | `[]` |
| `cert-manager.enable` | Using [cert manager](https://github.com/jetstack/cert-manager) for provisioning the certificates for the webhook server. You can follow [the cert manager documentation](https://cert-manager.io/docs/installation/) to install it. | `false` |
| `cert-manager.secretName` | TLS secret for certificates for the `${NAME}-webhook-service.${NAMESPACE}.svc` | `""` |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart. For example,

```console
$ helm install edge-operator -f values.yaml .
```
> **Tip**: You can use the default [values.yaml](https://github.com/emqx/edge-operator/tree/main/deploy/charts/edge-operator/values.yaml)

## Contributing

This chart is maintained at [github.com/edge/edge-operator](https://github.com/emqx/edge-operator/tree/main/deploy/charts/edge-operator).
