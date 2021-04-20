package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/hamfist/k8s-http-auth/middleware"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

type fakeTokenReviewer struct {
	clientauthv1.TokenReviewInterface

	store map[string]*authv1.TokenReview
}

func (tr *fakeTokenReviewer) Create(ctx context.Context, tokenReview *authv1.TokenReview, opts metav1.CreateOptions) (*authv1.TokenReview, error) {
	log := logr.FromContextOrDiscard(ctx)

	key := tokenReviewKey(tokenReview)
	stored, ok := tr.store[key]
	if ok {
		log.Info("found token review with key", "key", key, "store_keys", strings.Join(tr.storeKeys(), ","))
		return stored, nil
	}

	no := apierrors.NewUnauthorized("no")
	log.Error(no, "no token review with key", "key", key, "store_keys", strings.Join(tr.storeKeys(), ","))
	return nil, no
}

func (tr *fakeTokenReviewer) storeKeys() []string {
	keys := []string{}
	for key := range tr.store {
		keys = append(keys, key)
	}
	return keys
}

func tokenReviewKey(tokenReview *authv1.TokenReview) string {
	return fmt.Sprintf("%s-%s", tokenReview.Spec.Token, strings.Join(tokenReview.Spec.Audiences, ";"))
}

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

	tr := &fakeTokenReviewer{
		store: map[string]*authv1.TokenReview{
			tokenReviewKey(goodTokenReview):       goodTokenReview,
			tokenReviewKey(disallowedTokenReview): disallowedTokenReview,
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
