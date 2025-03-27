window.onload = function() {
    // Build a system
    const ui = SwaggerUIBundle({
      urls: [
      {{range $index, $u := .URLs}}
        {
          name: "{{$u.Name}}",
          url: "{{$u.URL}}",
        },
      {{end}}
      ],
      syntaxHighlight: {{.SyntaxHighlight}},
      deepLinking: {{.DeepLinking}},
      docExpansion: "{{.DocExpansion}}",
      showExtensions: {{.ShowExtensions}},
      persistAuthorization: {{.PersistAuthorization}},
      dom_id: "#{{.DomID}}",
      validatorUrl: null,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset
      ],
      plugins: [
        SwaggerUIBundle.plugins.DownloadUrl
      ],
      layout: "StandaloneLayout"
    })
  
    {{if .OAuth}}
    ui.initOAuth({
      clientId: "{{.OAuth.ClientId}}",
      realm: "{{.OAuth.Realm}}",
      appName: "{{.OAuth.AppName}}"
    })
    {{end}}
  
    window.ui = ui
}