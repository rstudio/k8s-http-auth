package local_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/rstudio/k8s-http-auth/reviewer/local"
	"github.com/stretchr/testify/assert"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTokenReviewCreator(t *testing.T) {
	td, err := ioutil.TempDir("", "k8s-http-auth-*")
	assert.Nil(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	trc, err := local.New(td)
	assert.Nil(t, err)
	assert.NotNil(t, trc)

	trdb, err := local.NewDB(td)
	assert.Nil(t, err)
	assert.NotNil(t, trdb)

	goodToken := "radiator"
	goodAudiences := []string{"foot", "knee"}

	err = trdb.Put(goodToken, goodAudiences)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tr, err := trc.Create(ctx, &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     goodToken,
			Audiences: goodAudiences,
		},
	}, metav1.CreateOptions{})

	assert.Nil(t, err)
	assert.NotNil(t, tr)
}
