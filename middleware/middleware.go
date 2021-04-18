package middleware // import "github.com/hamfist/k8s-http-auth/middleware"

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

const (
	AuthStatusContextKey contextKey = "k8s-http-auth.middleware.status"
)

var (
	errNoClientID       = errors.New("no client id")
	errNotAuthenticated = errors.New("not authenticated")
)

type contextKey string

type Options struct {
	IDHeader  string
	Audiences []string
}

func New(reviewer clientauthv1.TokenReviewInterface, opts *Options) func(http.Handler) http.Handler {
	mw := &middleware{
		reviewer:       reviewer,
		clientIDHeader: client.DefaultIDHeader,
		audiences:      nil,
	}

	if opts != nil {
		if opts.IDHeader != "" {
			mw.clientIDHeader = opts.IDHeader
		}

		if opts.Audiences != nil && len(opts.Audiences) > 0 {
			mw.audiences = opts.Audiences
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			status, ok := mw.isAuthorized(w, req)
			if !ok {
				return
			}

			next.ServeHTTP(
				w,
				req.WithContext(
					context.WithValue(
						req.Context(), AuthStatusContextKey, status,
					),
				))
		})
	}
}

type middleware struct {
	reviewer       clientauthv1.TokenReviewInterface
	clientIDHeader string
	audiences      []string
}

func (mw *middleware) isAuthorized(w http.ResponseWriter, req *http.Request) (*authv1.TokenReviewStatus, bool) {
	log := logr.FromContextOrDiscard(req.Context())

	clientID := req.Header.Get(mw.clientIDHeader)
	if len(clientID) == 0 {
		log.Error(errNoClientID, "missing header", "header", mw.clientIDHeader)

		http.Error(w, fmt.Sprintf("missing %q header", mw.clientIDHeader), http.StatusUnauthorized)
		return nil, false
	}

	tr := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     clientID,
			Audiences: mw.audiences,
		},
	}

	log.Info("creating token review")

	authResult, err := mw.reviewer.Create(req.Context(), tr, metav1.CreateOptions{})
	if err != nil {
		log.Error(err, "failed to create token review")

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return nil, false
	}

	if !authResult.Status.Authenticated {
		log.Error(errors.Wrap(errNotAuthenticated, authResult.Status.Error), "not authenticated")

		http.Error(w, authResult.Status.Error, http.StatusUnauthorized)
		return nil, false
	}

	log.Info("authenticated")
	return &authResult.Status, true
}
