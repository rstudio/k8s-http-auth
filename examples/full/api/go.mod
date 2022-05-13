module github.com/rstudio/k8s-http-auth/examples/full/api

go 1.18

replace github.com/rstudio/k8s-http-auth => ./k8s-http-auth/

require (
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/gorilla/mux v1.8.0
	github.com/rstudio/k8s-http-auth v0.4.3
	go.uber.org/zap v1.21.0
)

require (
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
)
