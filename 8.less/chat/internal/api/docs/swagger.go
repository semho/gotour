package docs

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func ServeSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	swaggerPath := filepath.Join("pkg", "chat", "v1", "chat.swagger.json")
	http.ServeFile(w, r, swaggerPath)
}

func ServeSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	swaggerTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Chat Service API</title>
    <link rel="stylesheet" type="text/css" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui.min.css" >
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-bundle.min.js"> </script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.15.5/swagger-ui-standalone-preset.min.js"> </script>
</head>
<body>
<div id="swagger-ui"></div>
<script>
window.onload = function() {
    const ui = SwaggerUIBundle({
        url: "/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
        ],
        plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout",
        displayRequestDuration: true,
        filter: true,
        requestInterceptor: (request) => {
            const sessionId = localStorage.getItem('SessionId');
            if (sessionId) {
                request.headers['Session-Id'] = sessionId;
            }
            return request;
        },
        onComplete: function() {
            ui.preauthorizeApiKey("SessionId", localStorage.getItem('SessionId'));
        }
    });

    // Add listener to update localStorage when SessionId is changed
    ui.getSystem().on("authChange", function() {
        const auth = ui.getState().get("auth");
        const sessionId = auth.getIn(["authorized", "SessionId", "value"]);
        if (sessionId) {
            localStorage.setItem('SessionId', sessionId);
        } else {
            localStorage.removeItem('SessionId');
        }
    });
}
</script>
</body>
</html>
`
	tmpl, err := template.New("swagger").Parse(swaggerTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
