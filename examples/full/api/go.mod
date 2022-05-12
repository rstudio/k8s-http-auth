module github.com/rstudio/k8s-http-auth/examples/full/api

go 1.18

replace github.com/rstudio/k8s-http-auth => ./k8s-http-auth/

require (
	github.com/go-logr/logr v1.2.2
	github.com/go-logr/zapr v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/rstudio/k8s-http-auth v0.4.1
	go.uber.org/zap v1.19.0
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)
