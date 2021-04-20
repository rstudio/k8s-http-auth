// Kubernetes HTTP auth things with support for projected service
// account token auth.
package k8shttpauth

import (
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/hamfist/k8s-http-auth/middleware"
	"github.com/hamfist/k8s-http-auth/reviewer"
)

var (
	// NewClientInterface returns an interface for getting the
	// client ID and building http requests with the necessary
	// header.
	NewClientInterface = client.New

	// NewMiddleware returns a new Middleware for use with
	// http mux (router).
	NewMiddleware = middleware.New

	// NewMiddlewareFunc returns a new middleware Func for use with
	// http mux (router).
	NewMiddlewareFunc = middleware.NewFunc

	// NewReviewer returns a reviewer for general token review
	// needs.
	NewReviewer = reviewer.New
)

type ClientInterface = client.Interface
type Middleware = middleware.Middleware
type MiddlewareFunc = middleware.Func
type Reviewer = reviewer.Reviewer
