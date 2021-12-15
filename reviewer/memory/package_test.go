package memory_test

import (
	"testing"

	"github.com/rstudio/k8s-http-auth/reviewer"
	"github.com/rstudio/k8s-http-auth/reviewer/memory"
)

func TestPackageInterface(t *testing.T) {
	var _ reviewer.TokenReviewCreator = new(memory.TokenReviewCreator)
}
