package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		fmt.Println("Incremented Hits: ", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsReset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Store(0)
		fmt.Println("Reset: ", cfg.fileserverHits.Load())
		next.ServeHTTP(w, r)
	})
}

func main() {
	apiCfg := &apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./")))))

	mux.Handle("/app/assets/logo.png", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %d", apiCfg.fileserverHits.Load())))
		fmt.Println("Hits: ", apiCfg.fileserverHits.Load())
	})

	mux.Handle("/reset", apiCfg.middlewareMetricsReset(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Hits: %d", apiCfg.fileserverHits.Load())))
		fmt.Println("Hits: ", apiCfg.fileserverHits.Load())
	})))

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Println("Server started on port 8080")
	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
