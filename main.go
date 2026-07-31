package main

import (
	"net/http"
	"fmt"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}


func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}


func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	page := fmt.Sprintf(`
	<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    	<p>Chirpy has been visited %d times!</p>
  	</body>
	</html>
	`, cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}


func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}


func main() {
	apiCfg := &apiConfig{}
	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}


