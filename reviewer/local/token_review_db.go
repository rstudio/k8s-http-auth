package local

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterbourgon/diskv/v3"
	authv1 "k8s.io/api/authentication/v1"
)

var (
	errEmptyToken = errors.New("empty token")
)

// TokenReviewDB allows direct access to the diskv-based database
// used by the local TokenReviewCreator.
type TokenReviewDB struct {
	db *diskv.Diskv
}

func NewDB(basePath string) (*TokenReviewDB, error) {
	dbOpts := diskv.Options{
		BasePath: filepath.Join(basePath, "db"),
		TempDir:  filepath.Join(basePath, "tmp"),
	}

	for _, p := range []string{dbOpts.BasePath, dbOpts.TempDir} {
		if err := os.MkdirAll(p, 0755); err != nil {
			return nil, err
		}
	}

	return &TokenReviewDB{db: diskv.New(dbOpts)}, nil
}

func (trc *TokenReviewDB) Get(token string, audiences []string) (*authv1.TokenReview, error) {
	if token == "" {
		return nil, errEmptyToken
	}

	trBytes, err := trc.db.Read(trc.tokenReviewKey(token, audiences))
	if err != nil {
		return nil, err
	}

	ret := &authv1.TokenReview{}
	if err := json.Unmarshal(trBytes, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (trc *TokenReviewDB) Put(token string, audiences []string) error {
	if token == "" {
		return errEmptyToken
	}

	trBytes, err := json.Marshal(&authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     token,
			Audiences: audiences,
		},
	})
	if err != nil {
		return err
	}

	return trc.db.Write(trc.tokenReviewKey(token, audiences), trBytes)
}

func (trc *TokenReviewDB) tokenReviewKey(token string, audiences []string) string {
	key := strings.Join([]string{
		token,
		strings.Join(audiences, ","),
	}, ":")
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}
