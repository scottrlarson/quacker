// page_handler.go

package main

import (
	"fmt"
	"net/http"
)

func renderPage(w http.ResponseWriter, r *http.Request, title string, content string) {
	w.Write([]byte(fmt.Sprintf(`<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
		<title>%s - Quacker</title>
		<link rel="stylesheet" href="/assets/css/bootstrap.min.css">
		<script src="/assets/js/bootstrap.bundle.min.js"></script>
	</head>
	<body>
		<div class="container mt-5">
			%s
			<div class="card">
				<div class="card-body">
					%s
				</div>
			</div>
			<footer class="mt-4 text-center">
				<small>&copy; 2025 Matthew Reider - <a href="https://github.com/mreider">GitHub</a></small>
			</footer>
		</div>
	</body>
	</html>`, title, renderMenu(r), content)))
}
