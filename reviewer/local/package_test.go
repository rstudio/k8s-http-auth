package local_test

import (
	"testing"

	"github.com/hamfist/k8s-http-auth/reviewer"
	"github.com/hamfist/k8s-http-auth/reviewer/local"
)

func TestPackageInterface(t *testing.T) {
	var _ reviewer.TokenReviewCreator = new(local.TokenReviewCreator)
}
