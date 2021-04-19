# k8s-http-auth

Kubernetes HTTP auth things.

[![Go Reference](https://pkg.go.dev/badge/github.com/hamfist/k8s-http-auth.svg)](https://pkg.go.dev/github.com/hamfist/k8s-http-auth)

## middleware

HTTP middleware to implement [intra-cluster communication with
service account token volume
projection](https://learnk8s.io/microservices-authentication-kubernetes#inter-service-authentication-using-service-account-token-volume-projection).

## client library

Client library to build requests that include a header to interact
with services that use the middleware.

## examples

A [full example](./examples/full) is available that includes an
api service that accesses a backing db service.
