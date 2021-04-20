package reviewertest

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	authv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthv1 "k8s.io/client-go/kubernetes/typed/authentication/v1"
)

type FakeTokenReviewer struct {
	clientauthv1.TokenReviewInterface

	Store map[string]*authv1.TokenReview
}

func (tr *FakeTokenReviewer) Create(ctx context.Context, tokenReview *authv1.TokenReview, opts metav1.CreateOptions) (*authv1.TokenReview, error) {
	log := logr.FromContextOrDiscard(ctx)

	key := FakeTokenReviewKey(tokenReview)
	stored, ok := tr.Store[key]
	if ok {
		log.Info("found token review with key", "key", key, "store_keys", strings.Join(tr.storeKeys(), ","))
		return stored, nil
	}

	no := apierrors.NewUnauthorized("no")
	log.Error(no, "no token review with key", "key", key, "store_keys", strings.Join(tr.storeKeys(), ","))
	return nil, no
}

func (tr *FakeTokenReviewer) storeKeys() []string {
	keys := []string{}
	for key := range tr.Store {
		keys = append(keys, key)
	}
	return keys
}

func FakeTokenReviewKey(tokenReview *authv1.TokenReview) string {
	return fmt.Sprintf("%s-%s", tokenReview.Spec.Token, strings.Join(tokenReview.Spec.Audiences, ";"))
}
