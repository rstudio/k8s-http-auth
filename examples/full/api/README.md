# `k8s-http-auth` client api service example

This example API is configured to use auth based on a projected
service account token to access a [`db` service](../db/main.go)
running in a separate pod.

The service account token audience is configured at the deployment
level in [`../deployment.yaml`](../deployment.yaml). The audience
value specified there *must match* the audience used in
[`../db/main.go`](../db/main.go) when configuring the
`k8s-http-auth` middleware. In this example, the values are
hard-coded to match.
