package api

import (
	"net/http"
	"net/url"
	"strings"

	swaggerdocs "robot-center/apps/server/internal/api/swaggerdocs"
)

const swaggerUIHTML = `<!doctype html>
<html lang="ko">
<head>
  <meta charset="utf-8">
  <title>Robot Center API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #f7f7f7; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/doc.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`

func (s *Server) handleSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func (s *Server) handleOpenAPIJSON(w http.ResponseWriter, _ *http.Request) {
	s.writeSwaggerDocJSON(w)
}

func (s *Server) handleSwaggerDocJSON(w http.ResponseWriter, _ *http.Request) {
	s.writeSwaggerDocJSON(w)
}

func (s *Server) writeSwaggerDocJSON(w http.ResponseWriter) {
	s.applySwaggerServerInfo()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerdocs.SwaggerInfo.ReadDoc()))
}

func (s *Server) applySwaggerServerInfo() {
	publicURL, err := url.Parse(strings.TrimSpace(s.config.AppServerPublicURL))
	if err != nil || publicURL.Host == "" {
		swaggerdocs.SwaggerInfo.Host = ""
		swaggerdocs.SwaggerInfo.Schemes = []string{}
		return
	}
	swaggerdocs.SwaggerInfo.Host = publicURL.Host
	if publicURL.Scheme != "" {
		swaggerdocs.SwaggerInfo.Schemes = []string{publicURL.Scheme}
	}
}
