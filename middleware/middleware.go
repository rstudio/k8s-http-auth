// Kubernetes HTTP auth middleware for managing access via client
// ID (service account token) present in request header.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/hamfist/k8s-http-auth/reviewer"
	"github.com/pkg/errors"
	clientauthv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

const (
	AuthStatusContextKey contextKey = "k8s-http-auth.middleware.status"
)

var (
	JSONNotImplementedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "not implemented",
			"details": strings.Join([]string{
				"you are likely seeing this because there is no \"next\" handler",
				"available in the k8s-http-auth middleware",
			}, " "),
		})
	})

	JSONUnauthorizedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	})

	errNilNext    = errors.New("nil next http.Handler")
	errNoClientID = errors.New("no client id")
)

type contextKey string

// Func is the function type returned from New for use as an
// http middleware.
type Func func(http.Handler) http.Handler

// Middleware is an http.Handler that knows how to wrap another
// http.Handler as in a middleware chain.
type Middleware interface {
	http.Handler

	// WithNext accepts the "next" http.Handler in the middleware
	// stack and returns the wrapping Middleware.
	WithNext(http.Handler) Middleware
}

// Options may be passed to New when creating a middleware
// func type.
type Options struct {
	// Audiences are passed directly with a token review when
	// validating a request.
	Audiences []string

	// IDHeader is the header key checked when validating a
	// request.
	IDHeader string

	// UnauthorizedHandler is used when the incoming request is not
	// authorized. The handler is expected to run
	// http.ResponseWriter.WriteHeader. If not provided, the
	// default will be JSONUnauthorizedHandler.
	UnauthorizedHandler http.Handler
}

// NewFunc creates a new Func for use with an http mux (router).
func NewFunc(rev clientauthv1.TokenReviewInterface, opts *Options) Func {
	return func(next http.Handler) http.Handler {
		return New(rev, opts).WithNext(next)
	}
}

// New creates a Middleware for use with an http mux (router).
func New(rev clientauthv1.TokenReviewInterface, opts *Options) Middleware {
	var revOpts *reviewer.Options = nil

	if opts != nil {
		revOpts = &reviewer.Options{Audiences: opts.Audiences}

		if opts.UnauthorizedHandler == nil {
			opts.UnauthorizedHandler = JSONUnauthorizedHandler
		}

		if opts.IDHeader == "" {
			opts.IDHeader = client.DefaultIDHeader
		}
	}

	return &middleware{
		rev:      reviewer.New(rev, revOpts),
		unAuthd:  opts.UnauthorizedHandler,
		idHeader: opts.IDHeader,
	}
}

type middleware struct {
	rev      reviewer.Reviewer
	next     http.Handler
	unAuthd  http.Handler
	idHeader string
}

// WithNext satisfies the HandlerWithNexter interface, returning
// this Middleware
func (mw *middleware) WithNext(next http.Handler) Middleware {
	mw.next = next
	return mw
}

// ServeHTTP satisfies the http.Handler interface
func (mw *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log := logr.FromContextOrDiscard(req.Context())

	clientID := req.Header.Get(mw.idHeader)
	if len(clientID) == 0 {
		log.Error(errNoClientID, "request is missing client id header", "key", mw.idHeader)
		mw.unAuthd.ServeHTTP(w, req)
		return
	}

	if err := mw.rev.Review(req.Context(), clientID); err != nil {
		if errors.Is(err, reviewer.ErrNotAuthenticated) {
			log.Error(err, "token review rejected")
		} else {
			log.Error(err, "token review failed")
		}
		mw.unAuthd.ServeHTTP(w, req)
		return
	}

	next := mw.next
	if next == nil {
		next = JSONNotImplementedHandler
	}

	next.ServeHTTP(
		w,
		req.WithContext(
			context.WithValue(
				req.Context(), AuthStatusContextKey, true,
			),
		),
	)
}
