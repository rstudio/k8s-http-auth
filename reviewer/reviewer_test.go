package reviewer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/rstudio/k8s-http-auth/reviewer"
	"github.com/rstudio/k8s-http-auth/reviewer/memory"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authentication/v1"
)

func TestReviewer(t *testing.T) {
	goodToken := "nancy"
	badToken := "nanobot"
	disallowedToken := "nan"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logrusr.NewLogger(logrus.New()).WithName("k8s-http-auth-test")
	ctx = logr.NewContext(ctx, log)

	goodTokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     goodToken,
			Audiences: []string{"hats"},
		},
		Status: authv1.TokenReviewStatus{
			Authenticated: true,
		},
	}

	disallowedTokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     disallowedToken,
			Audiences: []string{"hats"},
		},
		Status: authv1.TokenReviewStatus{
			Authenticated: false,
			Error:         "disallowed nan",
		},
	}

	tr := memory.New(
		goodTokenReview,
		disallowedTokenReview,
	)

	rev := reviewer.New(tr, &reviewer.Options{Audiences: []string{"hats"}})

	err := rev.Review(ctx, goodToken)
	assert.Nilf(t, err, "good token is allowed and results in nil error")

	err = rev.Review(ctx, disallowedToken)
	assert.NotNilf(t, err, "disallowed token is not allowed and results in non-nil error")
	assert.Truef(t, errors.Is(err, reviewer.ErrNotAuthenticated), "disallowed token error matches expected")

	err = rev.Review(ctx, badToken)
	assert.NotNilf(t, err, "bad token is not allowed and results in non-nil error")
	assert.Falsef(t, errors.Is(err, reviewer.ErrNotAuthenticated), "bad token error matches expected")
}
