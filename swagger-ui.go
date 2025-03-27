package doc

import (
	_ "embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
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

// SwaggerUIConfig stores swagger configuration variables.
type SwaggerUIConfig struct {
	Title            string
	DocDir           string
	PathPrefix       string
	UseIndexTemplate bool
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `mockedSwag.json`.
	URLs                 []DefinitionURL
	DocExpansion         string
	ShowExtensions       bool
	DomID                string
	DeepLinking          bool
	PersistAuthorization bool
	SyntaxHighlight      bool

	// The information for OAuth2 integration, if any.
	OAuth *OAuthConfig
}

type DefinitionURL struct {
	Name string
	URL  string
}

// OAuthConfig stores configuration for Swagger UI OAuth2 integration. See
// https://swagger.io/docs/open-source-tools/swagger-ui/usage/oauth2/ for further details.
type OAuthConfig struct {
	// The ID of the client sent to the OAuth2 IAM provider.
	ClientId string

	// The OAuth2 realm that the client should operate in. If not applicable, use empty string.
	Realm string

	// The name to display for the application in the authentication popup.
	AppName string
}

// WithTitle the page tile display on browser. default: API Doc
func WithTitle(title string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.Title = title
	}
}

func WithDocDir(dir string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.DocDir = dir
	}
}

func WithPathPrefix(prefix string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.PathPrefix = prefix
	}
}

func WithIndexTemplate(use bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.UseIndexTemplate = use
	}
}

// WithURL presents the url pointing to API definition (normally swagger.json or swagger.yaml).
func WithURL(url string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.URLs = append(c.URLs, DefinitionURL{URL: url, Name: url})
	}
}

// WithDefinitionURL presents the url pointing to API definition with name.
func WithDefinitionURL(du DefinitionURL) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.URLs = append(c.URLs, du)
	}
}

// WithDeepLinking true, false.
func WithDeepLinking(deepLinking bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.DeepLinking = deepLinking
	}
}

// WithSyntaxHighlight true, false.
func WithSyntaxHighlight(syntaxHighlight bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.SyntaxHighlight = syntaxHighlight
	}
}

// WithDocExpansion list, full, none.
func WithDocExpansion(docExpansion string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.DocExpansion = docExpansion
	}
}

// WithShowExtensions true, false.
func WithShowExtensions(show bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.ShowExtensions = show
	}
}

// WithDomID #swagger-ui.
func WithDomID(domID string) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.DomID = domID
	}
}

// WithPersistAuthorization Persist authorization information over browser close/refresh.
// Defaults to false.
func WithPersistAuthorization(enable bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.PersistAuthorization = enable
	}
}

func WithOAuth(config *OAuthConfig) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.OAuth = config
	}
}

func fixURLName() func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		for i := range c.URLs {
			if c.URLs[i].Name == "" {
				c.URLs[i].Name = c.URLs[i].URL
			}
		}
	}
}

func newConfig(configFns ...func(config *SwaggerUIConfig)) *SwaggerUIConfig {
	config := SwaggerUIConfig{
		Title:                "",
		DocDir:               "docs",
		URLs:                 []DefinitionURL{},
		DocExpansion:         "list",
		ShowExtensions:       true,
		DomID:                "swagger-ui",
		DeepLinking:          true,
		PersistAuthorization: false,
		SyntaxHighlight:      true,
	}

	for _, fn := range append(configFns, fixURLName()) {
		fn(&config)
	}

	if len(config.URLs) == 0 {
		addURLs := func(dir, path string) (err error) {
			relPath, _ := filepath.Rel(dir, path)
			if relPath == "" {
				relPath = path
			}
			switch filepath.Ext(relPath) {
			case ".json", ".yaml", ".yml":
				config.URLs = append(config.URLs, DefinitionURL{
					Name: filepath.Base(relPath),
					URL:  relPath,
				})
			}
			return nil
		}
		_ = filepath.WalkDir(config.DocDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				var realDir string
				realDir, err = os.Readlink(path)
				if err != nil {
					return err
				}
				err = filepath.WalkDir(realDir, func(path string, d fs.DirEntry, err error) error {
					return addURLs(realDir, path)
				})
			}
			return addURLs(config.DocDir, path)
		})
	}

	return &config
}

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
			if config.UseIndexTemplate {
				_ = indexHtml.Execute(w, config)
			} else {
				r.URL.Path = "/"
				docFS.ServeHTTP(w, r)
			}
		case "swagger-initializer.js":
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
