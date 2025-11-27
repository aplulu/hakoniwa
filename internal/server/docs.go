package server

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/aplulu/hakoniwa/api/openapi"
	"github.com/aplulu/hakoniwa/internal/config"
)

//go:embed html/swagger-ui.html
var swaggerUIHTML []byte

func serveSwaggerUI(w http.ResponseWriter, req *http.Request) {
	tmpl, err := template.New("swagger-ui").Parse(string(swaggerUIHTML))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, map[string]any{"SwaggerUIAuthorizeButtonEnabled": false}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveOpenAPI(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	if _, err := w.Write(openapi.GetOpenAPIAISchema()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newDocsHandler() http.Handler {
	mux := http.NewServeMux()
	if config.SwaggerUIEnabled() {
		mux.HandleFunc("/", serveSwaggerUI)
		mux.HandleFunc("/openapi.yaml", serveOpenAPI)
	}
	return mux
}
