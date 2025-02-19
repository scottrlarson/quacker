// site_list_handler.go

package main

import (
	"fmt"
	"net/http"
	"strings"
)

func siteListPage(w http.ResponseWriter, r *http.Request) {
	// Get logged-in user from the session cookie
	cookie, err := r.Cookie("quacker_user")
	if err != nil || cookie.Value == "" {
		// If user is not logged in, show login button instead of an error page
		content := `<div class="card">
			<div class="card-body text-center">
				<h1 class="mb-4">Welcome to Quacker</h1>
				<p>Please log in to manage your sites.</p>
				<a href="/login/github" class="btn btn-primary">Login with GitHub</a>
			</div>
		</div>`
		renderPage(w, r, "Login Required", content)
		return
	}
	loggedInUser := cookie.Value

	// Fetch sites for the logged-in user only
	sites, _ := rdb.Keys(ctx, "user_sites:"+loggedInUser+":*").Result()

	errorMessage := r.URL.Query().Get("error")
	successMessage := r.URL.Query().Get("success")

	content := `<script>
		function confirmDelete(domain) {
			if (confirm('Are you sure you want to delete this site? This action cannot be undone.')) {
				window.location.href = '/deletesite?domain=' + domain;
			}
		}
	</script>
	<div class="card">
		<div class="card-body">
			<h1 class="mb-4">` + loggedInUser + `&#39;s Sites</h1>
			<a href="/addsite" class="btn btn-primary mb-4">Add New Site</a>`

	if errorMessage != "" {
		content += `<div class="alert alert-danger" role="alert">` + errorMessage + `</div>`
	}

	if successMessage != "" {
		content += `<div class="alert alert-success" role="alert">` + successMessage + `</div>`
	}

	content += `
			<table class="table table-striped">
				<thead>
					<tr>
						<th>Domain</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>`

	if len(sites) == 0 {
		content += `<tr><td colspan="2">No sites available</td></tr>`
	} else {
		for _, siteKey := range sites {
			domain := strings.TrimPrefix(siteKey, "user_sites:"+loggedInUser+":")
			content += fmt.Sprintf(`<tr>
				<td>%s</td>
				<td>
					<a href="/js/%s" class="btn btn-secondary btn-sm">View HTML</a>
					<a href="/editsite?domain=%s" class="btn btn-secondary btn-sm">Edit</a>
					<a href="/downloadsubscribers?domain=%s" class="btn btn-secondary btn-sm">Subscribers</a>
					<button class="btn btn-danger btn-sm" onclick="confirmDelete('%s')">Delete</button>
				</td>
			</tr>`, domain, domain, domain, domain, domain)
		}
	}

	content += `
				</tbody>
			</table>
		</div>
	</div>`

	renderPage(w, r, "Your Sites", content)
}
