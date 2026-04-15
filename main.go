package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"redis/internal/api"
	ownmap "redis/internal/map"
)

func loadEnv(path string) map[string]string {
	env := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return env
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return env
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	env := loadEnv(".env")
	var om ownmap.Map
	switch env["MAP_TYPE"] {
	case "std":
		om = ownmap.NewStdMap()
	default:
		om = ownmap.NewOwnMap(1_000_000)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/set", logging(api.SetHandler(om)))
	mux.HandleFunc("/get", logging(api.GetHandler(om)))
	mux.HandleFunc("/docs", api.DocsHandler())
	mux.HandleFunc("/swagger.yaml", api.SwaggerHandler())

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("starting on :8080 with %d CPUs", runtime.NumCPU())
	log.Fatal(server.ListenAndServe())
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}
