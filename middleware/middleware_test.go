package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/hamfist/k8s-http-auth/middleware"
	"github.com/hamfist/k8s-http-auth/reviewer/reviewertest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authentication/v1"
	clientauthv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

func TestMiddleware(t *testing.T) {
	goodToken := "alfredo"
	badToken := "alfresco"
	disallowedToken := "alf"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logrusr.NewLogger(logrus.New()).WithName("k8s-http-auth-test")
	ctx = logr.NewContext(ctx, log)

	goodTokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     goodToken,
			Audiences: []string{"birbs"},
		},
		Status: authv1.TokenReviewStatus{
			Authenticated: true,
		},
	}

	disallowedTokenReview := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     disallowedToken,
			Audiences: []string{"birbs"},
		},
		Status: authv1.TokenReviewStatus{
			Authenticated: false,
			Error:         "disallowed alf",
		},
	}

	tr := &reviewertest.FakeTokenReviewer{
		Store: map[string]*authv1.TokenReview{
			reviewertest.FakeTokenReviewKey(goodTokenReview):       goodTokenReview,
			reviewertest.FakeTokenReviewKey(disallowedTokenReview): disallowedTokenReview,
		},
	}

	type tcFunc func(http.Handler, clientauthv1.TokenReviewInterface, *middleware.Options) http.Handler

	for _, f := range []tcFunc{
		func(next http.Handler, tr clientauthv1.TokenReviewInterface, opts *middleware.Options) http.Handler {
			return middleware.New(tr, opts).WithNext(next)
		},
		func(next http.Handler, tr clientauthv1.TokenReviewInterface, opts *middleware.Options) http.Handler {
			return middleware.NewFunc(tr, opts)(next)
		},
	} {
		state := struct {
			requests []string
		}{
			requests: []string{},
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			state.requests = append(state.requests, req.RequestURI)
			w.WriteHeader(http.StatusOK)
		})

		mw := f(next, tr, &middleware.Options{Audiences: []string{"birbs"}})
		req := httptest.NewRequest("GET", "/whatever", nil)
		w := httptest.NewRecorder()

		req.Header.Set(client.DefaultIDHeader, goodToken)

		mw.ServeHTTP(w, req.WithContext(ctx))

		assert.Len(t, state.requests, 1)
		assert.Equal(t, http.StatusOK, w.Code)

		req = httptest.NewRequest("GET", "/whatever", nil)
		w = httptest.NewRecorder()

		req.Header.Set(client.DefaultIDHeader, badToken)

		mw.ServeHTTP(w, req.WithContext(ctx))

		assert.Len(t, state.requests, 1)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		req = httptest.NewRequest("GET", "/whatever", nil)
		w = httptest.NewRecorder()

		req.Header.Set(client.DefaultIDHeader, disallowedToken)

		mw.ServeHTTP(w, req.WithContext(ctx))

		assert.Len(t, state.requests, 1)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}
