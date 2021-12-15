# `k8s-http-auth` full example

This is a full, albeit extremely simplified, example of how to use
the `k8s-http-auth` client library and middleware. The example uses
[gorilla/mux](https://pkg.go.dev/github.com/gorilla/mux) when
adding auth middleware, but any framework that accepts or can be
adapted to accept [this function
type](https://pkg.go.dev/github.com/rstudio/k8s-http-auth/middleware#Func)
should work.

## Running Locally

To run locally with [kind](https://kind.sigs.k8s.io/) via `make`:

```bash
# in this `examples/full` directory
make
```

The default `make` (`make all`) target will run the following targets:

- `build` - build the api and db images
- `kind-ensure-cluster` - ensure a kind cluster is available
- `kind-load` - load the api and db images into the kind cluster
- `apply` - apply the `./deployment.yaml` to the cluster
- `wait-for-db` - wait for the db pod to be ready
- `wait-for-api` - wait for the api pod to be ready

## Interacting Locally

A healthy api service with fully-functioning auth based on
projected service account token will respond with:

```
{"ok":"yep"}
```

## Cleaning up

The example may be removed from your kind cluster by running:

```bash
make delete
```
