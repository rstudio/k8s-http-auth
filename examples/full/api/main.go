package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/hamfist/k8s-http-auth/client"
	"github.com/sirupsen/logrus"
)

var (
	errDBUnavailable = errors.New("db unavailable")
)

func main() {
	log := logrusr.NewLogger(logrus.New()).WithName("k8s-http-auth-example-api")
	router := mux.NewRouter()
	ac := client.New(nil)

	addr := ":8080"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
		log.Info("using addr from ADDR", "addr", addr)
	}

	dbAddr := "http://db.k8s-http-auth-system:9090"
	if v := os.Getenv("DB_ADDR"); v != "" {
		dbAddr = v
		log.Info("using db addr from DB_ADDR", "db_addr", dbAddr)
	}

	router.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, "ok\n"); err != nil {
			log.Error(err, "failed to write response")
		}
	})

	router.HandleFunc("/", buildHome(ac, dbAddr))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Info(fmt.Sprintf("%s %s", req.Method, req.RequestURI))
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, req.WithContext(logr.NewContext(req.Context(), log)))
		})
	})

	log.Info("listening", "addr", addr)
	http.ListenAndServe(addr, router)
}

func buildHome(ac client.Interface, dbAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log := logr.FromContextOrDiscard(req.Context())

		req, err := http.NewRequest("GET", dbAddr, nil)
		if err != nil {
			log.Error(err, "failed to build db request")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"err": err.Error()})
			return
		}

		req, err = ac.WithHeader(req)
		if err != nil {
			log.Error(err, "failed to get request with client id header")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"err": err.Error()})
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error(err, "failed to get db state")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"err": err.Error()})
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Error(errDBUnavailable, "failed to get db state", "status", resp.StatusCode)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"err": fmt.Sprintf("failed to get db state; status=%v", resp.StatusCode),
			})
			return
		}

		dbState := struct {
			OK     string `json:"ok"`
			UTCNow string `json:"utcnow"`
		}{}

		if err := json.NewDecoder(resp.Body).Decode(&dbState); err != nil {
			log.Error(err, "failed to decode db state")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"err": err.Error()})
			return
		}

		log.Info("fetched db state", "state", dbState)

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(map[string]string{"ok": dbState.OK}); err != nil {
			log.Error(err, "failed to json encode response")
		}
	}
}
