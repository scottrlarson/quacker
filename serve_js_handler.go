// serve_js_handler.go

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func serveJS(w http.ResponseWriter, r *http.Request) {
	domain := mux.Vars(r)["domain"]
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		http.Error(w, "Configuration not found", http.StatusInternalServerError)
		return
	}
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		http.Error(w, "Invalid configuration format", http.StatusInternalServerError)
		return
	}

	// Check if the site exists using full domain key
	loggedInUser, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	fullDomainKey := "user_sites:" + loggedInUser.Value + ":" + domain
	if rdb.Get(ctx, fullDomainKey).Err() != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	formCode := fmt.Sprintf(`
<form id="subscribe-form" action="https://%s/subscribe" method="POST">
    <input type="hidden" name="owner" value="%s">
    <input type="hidden" name="domain" value="%s">
	<div class="d-flex align-items-center gap-2" id="form-content">
				<input type="email" class="form-control form-control-sm" id="email" name="email" placeholder="email" required>
				<button type="submit" class="btn btn-secondary btn-sm">Subscribe</button>
	</div>
	</form>
	<div id="subscribe-message" class="mt-3"></div>

<script>
		document.getElementById('subscribe-form').addEventListener('submit', function(event) {
			event.preventDefault();
			
			var form = document.getElementById('subscribe-form');
			var messageDiv = document.getElementById('subscribe-message');

			// Hide the form using Bootstrap's d-none class
			form.classList.add('d-none');

			// Clear previous messages
			messageDiv.innerHTML = '';

			var formData = new FormData(form);
			fetch(form.action, {
				method: 'POST',
				body: formData
			})
			.then(response => {
				if (response.ok) {
					return response.text();
				}
				throw new Error('Subscription failed');
			})
			.then(message => {
				messageDiv.innerHTML = '<div class="alert alert-success">' + message + '</div>';
				setTimeout(() => {
					messageDiv.innerHTML = '';  // Clear message
					form.classList.remove('d-none'); // Show form again
					form.reset();
				}, 3000);
			})
			.catch(error => {
				messageDiv.innerHTML = '<div class="alert alert-danger">' + error.message + '</div>';
				setTimeout(() => {
					messageDiv.innerHTML = '';  // Clear message
					form.classList.remove('d-none'); // Show form again
				}, 3000);
			});
		});
	</script>
	`, config.Hostname, loggedInUser.Value, domain)

	content := fmt.Sprintf(`
		<h2>Copy and Paste This Form to Your Static Site</h2>
		<p>Add this subscription form to your static site. When users enter their email, they will be subscribed to updates.</p>
		<div class="mb-3">
			<textarea id="form-code" class="form-control" rows="5" style="font-size: 12px;" readonly>%s</textarea>
		</div>
		<button class="btn btn-primary" onclick="navigator.clipboard.writeText(document.getElementById('form-code').value)">Copy to Clipboard</button>
	`, formCode)

	renderPage(w, r, "Generate Subscription Form", content)
}
