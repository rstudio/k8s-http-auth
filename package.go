// Kubernetes HTTP auth things with support for projected service
// account token auth.
package k8shttpauth

import (
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/hamfist/k8s-http-auth/middleware"
)

var (
	// NewMiddleware returns a new Middleware for use with
	// http mux (router).
	NewMiddleware = middleware.New

	// NewMiddlewareFunc returns a new middleware Func for use with
	// http mux (router).
	NewMiddlewareFunc = middleware.NewFunc

	// NewClientInterface returns an interface for getting the
	// client ID and building http requests with the necessary
	// header.
	NewClientInterface = client.New
)

type Client = client.Interface
type Middleware = middleware.HandlerWithNexter
type MiddlewareFunc = middleware.Func
