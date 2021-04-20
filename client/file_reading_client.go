package client

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

type fileReadingClient struct {
	loopInterval time.Duration
	path         string
	header       string
	cachedToken  string
	cachedError  error
}

func (frc *fileReadingClient) started(ctx context.Context) Interface {
	if frc.loopInterval <= 0 {
		frc.loopInterval = 1 * time.Hour
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := frc.refreshToken(); err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			time.Sleep(frc.loopInterval)
		}
	}()

	return frc
}

func (frc *fileReadingClient) refreshToken() error {
	tokenBytes, err := ioutil.ReadFile(frc.path)
	if err != nil {
		frc.cachedError = err
		return err
	}

	frc.cachedToken = strings.TrimSpace(string(tokenBytes))
	frc.cachedError = nil
	return nil
}

func (frc *fileReadingClient) ID(ctx context.Context) (string, error) {
	log := logr.FromContextOrDiscard(ctx)

	if frc.cachedToken != "" {
		log.Info("returning cached client id token")
		return frc.cachedToken, nil
	}

	log.Error(frc.cachedError, "failed to read client id token")
	return "", frc.cachedError
}

func (frc *fileReadingClient) WithHeader(req *http.Request) (*http.Request, error) {
	return withHeader(frc, req, frc.header)
}
