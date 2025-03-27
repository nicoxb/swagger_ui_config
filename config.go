package swagger_ui_config

import (
	"io/fs"
	"os"
	"path/filepath"
)

// SwaggerUIConfig stores swagger configuration variables.
type SwaggerUIConfig struct {
	Title                string
	DocDir               string
	PathPrefix           string
	DisableIndexTemplate bool
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

func WithDisableIndexTemplate(disable bool) func(*SwaggerUIConfig) {
	return func(c *SwaggerUIConfig) {
		c.DisableIndexTemplate = disable
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
