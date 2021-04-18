package client // import "github.com/hamfist/k8s-http-auth/client"

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

type expiringTokenReader struct {
	loopInterval time.Duration
	path         string
	header       string
	cachedToken  string
	cachedError  error
	loopCtx      context.Context
	loopCancel   context.CancelFunc
}

func (tr *expiringTokenReader) started() Interface {
	tr.loopCtx, tr.loopCancel = context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-tr.loopCtx.Done():
				return
			default:
			}

			if err := tr.refreshToken(); err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			time.Sleep(tr.loopInterval)
		}
	}()

	return tr
}

func (tr *expiringTokenReader) refreshToken() error {
	tokenBytes, err := ioutil.ReadFile(tr.path)
	if err != nil {
		tr.cachedError = err
		return err
	}

	tr.cachedToken = strings.TrimSpace(string(tokenBytes))
	tr.cachedError = nil
	return nil
}

func (tr *expiringTokenReader) ID(ctx context.Context) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	select {
	case <-ctx.Done():
		tr.loopCancel()
		log.Info("stopping client id token reader loop due to context done")
		return "", ctx.Err()
	default:
	}

	if tr.cachedToken != "" {
		log.Info("returning cached client id token")
		return tr.cachedToken, nil
	}

	log.Error(tr.cachedError, "failed to read client id token")
	return "", tr.cachedError
}

func (tr *expiringTokenReader) WithHeader(req *http.Request) (*http.Request, error) {
	return withHeader(tr, req, tr.header)
}
