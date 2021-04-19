package client // import "github.com/hamfist/k8s-http-auth/client"

import (
	"context"
	"net/http"
	"time"
)

const (
	DefaultIDHeader = "X-Client-Id"
)

var (
	LongLivedTokenOptions = &Options{
		IDHeader:    DefaultIDHeader,
		TokenPath:   "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TokenExpiry: -1 * time.Second,
	}

	ExpiringTokenOptions = &Options{
		IDHeader:    DefaultIDHeader,
		TokenPath:   "/var/run/secrets/tokens/api-token",
		TokenExpiry: 5 * time.Minute,
	}
)

type Interface interface {
	ID(context.Context) (string, error)
	WithHeader(*http.Request) (*http.Request, error)
}

type Options struct {
	IDHeader    string
	TokenPath   string
	TokenExpiry time.Duration
}

func New(opts *Options) Interface {
	if opts == nil {
		opts = ExpiringTokenOptions
	}

	if opts.TokenExpiry < 0 {
		return &longLivedTokenReader{
			path:   opts.TokenPath,
			header: opts.IDHeader,
		}
	}

	etr := &expiringTokenReader{
		loopInterval: opts.TokenExpiry,
		path:         opts.TokenPath,
		header:       opts.IDHeader,
	}
	return etr.started()
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
