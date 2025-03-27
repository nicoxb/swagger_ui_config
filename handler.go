package swagger_ui_config

import (
	_ "embed"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	swaggerFiles "github.com/swaggo/files/v2"
)

//go:embed index.tpl.html
var indexTplHtmlFile string

//go:embed index.tpl.js
var indexTplJSFile string

//go:embed index.tpl.css
var indexTplCSSFile string

func Handler(options ...func(*SwaggerUIConfig)) http.HandlerFunc {
	config := newConfig(options...)

	indexHtml := template.Must(template.New("swagger-index.html").Parse(indexTplHtmlFile))
	indexJS := template.Must(template.New("swagger-index.js").Parse(indexTplJSFile))
	indexCSS := template.Must(template.New("swagger-index.css").Parse(indexTplCSSFile))

	docFS := http.FileServer(http.Dir(config.DocDir))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if f, ok := w.(http.Flusher); ok {
			defer f.Flush()
		}

		reqPath := r.URL.EscapedPath()                             // full path
		prefix := "/" + strings.TrimPrefix(config.PathPrefix, "/") // ensure only one slash at path header
		relPath := strings.TrimPrefix(reqPath, prefix)             // relative path

		switch relPath {
		case "", "/", "index", "index.html", "/index.html":
			if !config.DisableIndexTemplate {
				_ = indexHtml.Execute(w, config)
			} else {
				r.URL.Path = "/"
				docFS.ServeHTTP(w, r)
			}
		case "/swagger-initializer.js":
			_ = indexJS.Execute(w, config)
		case "index.css":
			_ = indexCSS.Execute(w, config)
		default:
			r.URL.Path = relPath
			switch filepath.Ext(relPath) {
			case ".json", ".yaml", ".yml":
				// read file content directly
				docFS.ServeHTTP(w, r)
				return
			}
			http.FileServer(http.FS(swaggerFiles.FS)).ServeHTTP(w, r)
		}
	}
}
