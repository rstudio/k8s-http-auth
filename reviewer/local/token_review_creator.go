package local

import (
	"context"

	authv1 "k8s.io/api/authentication/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TokenReviewCreator fulfills the TokenReviewCreator
// interface that is expected to be typically fulfilled by
// k8s.io/client-go/kubernetes/typed/authentication/v1.TokenReviewInterface.
type TokenReviewCreator struct {
	db *TokenReviewDB
}

func New(basePath string) (*TokenReviewCreator, error) {
	db, err := NewDB(basePath)
	if err != nil {
		return nil, err
	}

	return &TokenReviewCreator{db: db}, nil
}

func (trc *TokenReviewCreator) Create(
	ctx context.Context,
	tr *authv1.TokenReview,
	_ metav1.CreateOptions,
) (*authv1.TokenReview, error) {
	ret, err := trc.db.Get(tr.Spec.Token, tr.Spec.Audiences)
	if err != nil {
		return nil, apierrors.NewUnauthorized(err.Error())
	}

	return ret, nil
}
