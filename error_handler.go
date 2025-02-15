// error_handler.go

package main

import (
	"fmt"
	"net/http"
)

func renderErrorPage(w http.ResponseWriter, r *http.Request, message string) {
	content := fmt.Sprintf(`
		<h1 class="mb-4">Error</h1>
		<p class="text-danger">%s</p>
	`, message)

	renderPage(w, r, "Error", content)
}
