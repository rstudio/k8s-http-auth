// Kubernetes HTTP auth client interface for reading client IDs and
// building requests.
package client

import (
	"context"
	"net/http"
	"time"
)

const (
	// DefaultIDHeader is the default header key used via
	// Interface.WithHeader. This constant is exported so that the
	// middleware can also use it as a default value.
	DefaultIDHeader = "X-Client-Id"
)

var (
	// LongLivedTokenOptions contains values that match the default
	// behavior for long-lived (no expiry) service account tokens.
	LongLivedTokenOptions = &Options{
		IDHeader:    DefaultIDHeader,
		TokenPath:   "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TokenExpiry: -1 * time.Second,
	}

	// ExpiringTokenOptions contains values that match those used
	// in the tutorial at
	// https://learnk8s.io/microservices-authentication-kubernetes.
	ExpiringTokenOptions = &Options{
		IDHeader:    DefaultIDHeader,
		TokenPath:   "/var/run/secrets/tokens/api-token",
		TokenExpiry: 5 * time.Minute,
	}
)

type IDer interface {
	// ID returns the client ID as typically defined in a service
	// account token.
	ID(context.Context) (string, error)
}

type WithHeaderer interface {
	// WithHeader returns a clone of the given request that
	// includes the client ID set as the value of the configured
	// header.
	WithHeader(*http.Request) (*http.Request, error)
}

// Interface provides methods for working with client IDs as
// provided by service account tokens.
type Interface interface {
	IDer
	WithHeaderer
}

// Options may be passed to New when creating a client Interface.
type Options struct {
	// IDHeader is the header key used when building a request
	// via Interface.WithHeader.
	IDHeader string

	// TokenPath is the file path from which the client ID will be
	// read via Interface.ID.
	TokenPath string

	// TokenExpiry defines the expected expiry time of the token
	// read from TokenPath. A TokenExpiry value that is less than
	// zero will build a long lived token reader that caches the
	// token value indefinitely.
	TokenExpiry time.Duration
}

// New creates a new client interface for use with building
// requests that contain the necessary auth headers. A nil value
// for the *Options argument is allowed and will result in the
// ExpiringTokenOptions being used.
func New(ctx context.Context, opts *Options) Interface {
	if opts == nil {
		opts = ExpiringTokenOptions
	}

	frc := &fileReadingClient{
		path:         opts.TokenPath,
		header:       opts.IDHeader,
		loopInterval: opts.TokenExpiry,
	}

	return frc.started(ctx)
}

func withHeader(tr Interface, req *http.Request, key string) (*http.Request, error) {
	tokenString, err := tr.ID(req.Context())
	if err != nil {
		return req, err
	}

	newReq := req.Clone(req.Context())
	newReq.Header.Set(key, tokenString)
	return newReq, nil
}
