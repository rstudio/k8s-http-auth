package client_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	td, err := ioutil.TempDir("", "k8s-http-auth.*")
	assert.Nil(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	tokenPath := filepath.Join(td, "api-token")

	cl := client.New(&client.Options{
		TokenExpiry: 10 * time.Minute,
		TokenPath:   tokenPath,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logrusr.NewLogger(logrus.New()).WithName("k8s-http-auth-test")
	ctx = logr.NewContext(ctx, log)

	time.Sleep(100 * time.Millisecond)

	tok, err := cl.ID(ctx)
	assert.NotNilf(t, err, "initial read fails because the file %q does not exist", tokenPath)

	err = ioutil.WriteFile(tokenPath, []byte("sturgeon\n"), 0640)
	assert.Nilf(t, err, "writing to token path %q resulted in an error", tokenPath)

	time.Sleep(200 * time.Millisecond)

	tok, err = cl.ID(ctx)
	assert.Nilf(t, err, "subsequent read succeeds because the file %q exists now", tokenPath)
	assert.Equalf(t, "sturgeon", tok, "token matched expected value")
}
