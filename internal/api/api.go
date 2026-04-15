package api

import (
	"encoding/json"
	"net/http"
	"time"

	ownmap "redis/internal/map"
)

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int    `json:"ttl_ms"`
}

type GetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func SetHandler(om ownmap.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		var expiresAt time.Time
		if req.TTL > 0 {
			expiresAt = time.Now().Add(time.Duration(req.TTL) * time.Millisecond)
		}
		om.Set(req.Key, req.Value, expiresAt)
		w.WriteHeader(http.StatusOK)
	}
}

func DocsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>API Docs</title>
	<meta charset="utf-8"/>
	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
	SwaggerUIBundle({ url: "/swagger.yaml", dom_id: '#swagger-ui' })
</script>
</body>
</html>`))
	}
}

func SwaggerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger.yaml")
	}
}
func GetHandler(om ownmap.Map) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}
		value := om.Get(key)
		if value == "" {
			http.Error(w, "key not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GetResponse{Key: key, Value: value})
	}
}
