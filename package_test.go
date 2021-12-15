package k8shttpauth_test

import (
	"testing"

	k8shttpauth "github.com/rstudio/k8s-http-auth"
)

func TestPackageInterface(t *testing.T) {
	_ = map[string]interface{}{
		"ClientInterface":    new(k8shttpauth.ClientInterface),
		"Middleware":         new(k8shttpauth.Middleware),
		"MiddlewareFunc":     new(k8shttpauth.MiddlewareFunc),
		"NewClientInterface": k8shttpauth.NewClientInterface,
		"NewMiddleware":      k8shttpauth.NewMiddleware,
		"NewMiddlewareFunc":  k8shttpauth.NewMiddlewareFunc,
		"NewReviewer":        k8shttpauth.NewReviewer,
		"Reviewer":           new(k8shttpauth.Reviewer),
	}
}
