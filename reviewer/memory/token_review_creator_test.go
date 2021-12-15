package memory_test

import (
	"context"
	"testing"

	"github.com/rstudio/k8s-http-auth/reviewer/memory"
	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTokenReviewCreator(t *testing.T) {
	goodToken := "tabletop"
	goodAudiences := []string{"hand", "cake"}

	goodTokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     goodToken,
			Audiences: goodAudiences,
		},
	}

	trc := memory.New(goodTokenReview)
	assert.NotNil(t, trc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr, err := trc.Create(ctx, goodTokenReview, metav1.CreateOptions{})

	assert.Nil(t, err)
	assert.NotNil(t, tr)
}
