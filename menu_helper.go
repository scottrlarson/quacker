package main

import "net/http"

func renderMenu(r *http.Request) string {
	_, err := r.Cookie("quacker_user")
	userLoggedIn := (err == nil)

	menu := `
	<nav class="navbar navbar-expand-lg navbar-light bg-light mb-4 px-3">
		<a class="navbar-brand" href="/">🦆 Quacker</a>
		<button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
			<span class="navbar-toggler-icon"></span>
		</button>
		<div class="collapse navbar-collapse" id="navbarNav">
			<ul class="navbar-nav mr-auto">
			</ul>
			<div class="d-flex ms-auto align-items-center">
	`

	if userLoggedIn {
		menu += `
				<a href="/sites" class="btn btn-outline-primary me-2">Sites</a>
				<a href="/logout" class="btn btn-outline-danger">Logout</a>
			`
	}

	menu += `
			</div>
		</div>
	</nav>`

	return menu
}