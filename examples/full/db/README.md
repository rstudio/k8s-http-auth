# `k8s-http-auth` db service example

This example db service is configured to restrict access to
requests that provide a `X-Client-Id` header (the default) with a
value that is verified via [token
review](https://pkg.go.dev/k8s.io/api/authentication/v1#TokenReview).

The binding to cluster role `system:auth-delegator` that allows
this service to create token reviews is configured at the
deployment level in [`../deployment.yaml`](../deployment.yaml).

The audiences allowed by the db service are hard-coded in this
example to `api-db`, which matches the service account token
audience specified for the [api service](../api/main.go).
