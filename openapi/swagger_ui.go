package openapi

import "embed"

//go:embed swagger_ui.html
var SwaggerUIHTML embed.FS

// SwaggerUIHandler returns the Swagger UI HTML content
func SwaggerUIHandler() ([]byte, error) {
	return SwaggerUIHTML.ReadFile("swagger_ui.html")
}
