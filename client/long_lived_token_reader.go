package client

import (
	"context"
	"io/ioutil"
	"net/http"
)

type longLivedTokenReader struct {
	cachedToken string
	header      string
	path        string
}

func (tr *longLivedTokenReader) ID(ctx context.Context) (string, error) {
	if len(tr.cachedToken) > 0 {
		return tr.cachedToken, nil
	}

	tokenBytes, err := ioutil.ReadFile(tr.path)
	if err != nil {
		return "", err
	}

	tr.cachedToken = string(tokenBytes)

	return tr.cachedToken, nil
}

func (tr *longLivedTokenReader) WithHeader(req *http.Request) (*http.Request, error) {
	return withHeader(tr, req, tr.header)
}
