package k8shttpauth_test

import (
	"testing"

	k8shttpauth "github.com/hamfist/k8s-http-auth"
)

func TestPackageInterface(t *testing.T) {
	_ = map[string]interface{}{
		"Client":             new(k8shttpauth.Client),
		"Middleware":         new(k8shttpauth.Middleware),
		"MiddlewareFunc":     new(k8shttpauth.MiddlewareFunc),
		"NewClientInterface": k8shttpauth.NewClientInterface,
		"NewMiddleware":      k8shttpauth.NewMiddleware,
		"NewMiddlewareFunc":  k8shttpauth.NewMiddlewareFunc,
	}
}
