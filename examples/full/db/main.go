package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/gorilla/mux"
	"github.com/rstudio/k8s-http-auth/middleware"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	zl, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to set up logger: %v\n", err)
		os.Exit(1)
	}

	log := zapr.NewLogger(zl).WithName("k8s-http-auth-example-api")
	router := mux.NewRouter()

	addr := ":9090"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
		log.Info("using addr from ADDR", "addr", addr)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "failed to get kubernetes cluster config")
		os.Exit(1)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "failed to get kubernetes client set")
		os.Exit(1)
	}

	router.HandleFunc("/", state)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Info(fmt.Sprintf("%s %s", req.Method, req.RequestURI))
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req.WithContext(logr.NewContext(req.Context(), log)))
		})
	})
	router.Use(mux.MiddlewareFunc(middleware.NewFunc(
		clientSet.AuthenticationV1().TokenReviews(),
		&middleware.Options{
			Audiences: []string{"api-db"},
		},
	)))

	log.Info("listening", "addr", addr)
	http.ListenAndServe(addr, router)
}

func state(w http.ResponseWriter, req *http.Request) {
	log := logr.FromContextOrDiscard(req.Context())

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"ok":     "yep",
		"utcnow": time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		log.Error(err, "failed to json encode response")
	}
}
