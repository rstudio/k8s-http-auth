// Kubernetes client id reviewer.
package reviewer

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

type Reviewer interface {
	Review(context.Context, string) error
}

// TokenReviewCreator is a subset of
// k8s.io/client-go/kubernetes/typed/authentication/v1.TokenReviewInterface
// that allows for easier interface fulfillment such as by
// github.com/rstudio/k8s-http-auth/reviewer/local.TokenReviewCreator
// and
// github.com/rstudio/k8s-http-auth/reviewer/memory.TokenReviewCreator.
// In typical production usage, this interface is satisfied by
// (*k8s.io/client-go/kubernetes.Clientset).AuthenticationV1().TokenReviews().
type TokenReviewCreator interface {
	Create(context.Context, *authv1.TokenReview, metav1.CreateOptions) (*authv1.TokenReview, error)
}

type Options struct {
	Audiences []string
}

func New(rev TokenReviewCreator, opts *Options) Reviewer {
	kcl := &k8sClientReviewer{
		rev:       rev,
		audiences: nil,
	}

	if opts != nil {
		if opts.Audiences != nil && len(opts.Audiences) > 0 {
			kcl.audiences = opts.Audiences
		}
	}

	return kcl
}

type k8sClientReviewer struct {
	rev       TokenReviewCreator
	audiences []string
}

func (kcl *k8sClientReviewer) Review(ctx context.Context, token string) error {
	log := logr.FromContextOrDiscard(ctx)

	tr := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     token,
			Audiences: kcl.audiences,
		},
	}

	log.Info("creating token review")

	authResult, err := kcl.rev.Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		log.Error(err, "failed to create token review")

		return err
	}

	if !authResult.Status.Authenticated {
		err := errors.Wrap(ErrNotAuthenticated, authResult.Status.Error)
		log.Error(err, "not authenticated")

		return err
	}

	log.Info("authenticated")
	return nil
}
