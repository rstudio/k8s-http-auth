// Kubernetes HTTP auth middleware for managing access via client
// ID (service account token) present in request header.
package middleware

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
	errNilNext          = errors.New("nil next http.Handler")
	errNoClientID       = errors.New("no client id")
	errNotAuthenticated = errors.New("not authenticated")
)

type contextKey string

// Func is the function type returned from New for use as an
// http middleware.
type Func func(http.Handler) http.Handler

// HandlerWithNexter is an http.Handler that knows how to wrap another
// http.Handler as in a middleware chain
type HandlerWithNexter interface {
	http.Handler

	// WithNext accepts the "next" http.Handler in the middleware
	// stack and returns the wrapping http.Handler.
	WithNext(http.Handler) http.Handler
}

// Options may be passed to New when creating a middleware
// func type.
type Options struct {
	// IDHeader is the header key checked when validating a
	// request.
	IDHeader string

	// Audiences are passed directly with a token review when
	// validating a request.
	Audiences []string
}

// NewFunc creates a new Func for use with an http mux (router).
func NewFunc(reviewer clientauthv1.TokenReviewInterface, opts *Options) Func {
	return func(next http.Handler) http.Handler {
		return New(reviewer, opts).WithNext(next)
	}
}

// New creates a HandlerWithNexter for use with an http mux (router).
func New(reviewer clientauthv1.TokenReviewInterface, opts *Options) HandlerWithNexter {
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

	return mw
}

type middleware struct {
	reviewer       clientauthv1.TokenReviewInterface
	clientIDHeader string
	audiences      []string
	next           http.Handler
}

// WithNext satisfies the HandlerWithNexter interface, returning this middleware
// as an http.Handler
func (mw *middleware) WithNext(next http.Handler) http.Handler {
	mw.next = next
	return mw
}

// ServeHTTP satisfies the http.Handler interface
func (mw *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log := logr.FromContextOrDiscard(req.Context())

	status, ok := mw.isAuthorized(w, req)
	if !ok {
		return
	}

	if mw.next == nil {
		log.Error(errNilNext, "likely because WithNext has not been called")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mw.next.ServeHTTP(
		w,
		req.WithContext(
			context.WithValue(
				req.Context(), AuthStatusContextKey, status,
			),
		),
	)
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
