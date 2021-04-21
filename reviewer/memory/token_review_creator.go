package memory

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	authv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TokenReviewCreator fulfills the TokenReviewCreator
// interface that is expected to be typically fulfilled by
// k8s.io/client-go/kubernetes/typed/authentication/v1.TokenReviewInterface.
type TokenReviewCreator struct {
	db map[string]*authv1.TokenReview
}

func New(entries ...*authv1.TokenReview) *TokenReviewCreator {
	trc := &TokenReviewCreator{db: map[string]*authv1.TokenReview{}}
	for _, tr := range entries {
		trc.db[trc.tokenReviewKey(tr)] = tr
	}
	return trc
}

func (trc *TokenReviewCreator) Create(ctx context.Context, tokenReview *authv1.TokenReview, opts metav1.CreateOptions) (*authv1.TokenReview, error) {
	log := logr.FromContextOrDiscard(ctx)

	key := trc.tokenReviewKey(tokenReview)
	stored, ok := trc.db[key]
	if ok {
		log.Info("found token review with key", "key", key, "store_keys", strings.Join(trc.storeKeys(), ","))
		return stored, nil
	}

	no := apierrors.NewUnauthorized("no")
	log.Error(no, "no token review with key", "key", key, "store_keys", strings.Join(trc.storeKeys(), ","))
	return nil, no
}

func (trc *TokenReviewCreator) storeKeys() []string {
	keys := []string{}
	for key := range trc.db {
		keys = append(keys, key)
	}
	return keys
}

func (trc *TokenReviewCreator) tokenReviewKey(tr *authv1.TokenReview) string {
	return fmt.Sprintf("%s-%s", tr.Spec.Token, strings.Join(tr.Spec.Audiences, ";"))
}
