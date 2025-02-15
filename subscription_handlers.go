// subscription_handlers.go

package main

import (
	"net/http"
	"net/url"
	"strings"
)

func subscribe(w http.ResponseWriter, r *http.Request) {
	// Allow CORS for requests coming from static blogs
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	email := r.FormValue("email")
	if !emailReg.MatchString(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}
	domain := r.FormValue("domain")
	owner := r.FormValue("owner")
	if domain == "" || owner == "" {
		renderErrorPage(w, r, "400 - Bad Request: Missing domain or owner parameter.")
		return
	}

	// Validate that the request originates from the same domain
	referer := r.Header.Get("Referer")
	if referer == "" {
		http.Error(w, "Missing referer header", http.StatusForbidden)
		return
	}

	parsedReferer, err := url.Parse(referer)
	if err != nil || !strings.HasSuffix(parsedReferer.Host, domain) {
		http.Error(w, "Unauthorized request origin", http.StatusForbidden)
		return
	}

	// Check if the site exists in Redis for the given owner
	siteExists, err := rdb.Exists(ctx, "user_sites:"+owner+":"+domain).Result()
	if err != nil || siteExists == 0 {
		renderErrorPage(w, r, "400 - Bad Request: No matching site found for owner.")
		return
	}

	// Check if email is already subscribed
	emailExists, _ := rdb.SIsMember(ctx, "subs:"+owner+":"+domain, email).Result()
	if emailExists {
		http.Error(w, "This email is already subscribed.", http.StatusConflict)
		return
	}

	// Store subscriber in Redis under owner-specific key
	rdb.SAdd(ctx, "subs:"+owner+":"+domain, email)

	w.Write([]byte("Success!"))
}

func unsubscribe(w http.ResponseWriter, r *http.Request) {
	// Allow CORS for requests coming from static blogs
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	email := r.FormValue("email")
	domain := r.FormValue("domain")
	owner := r.FormValue("owner")
	if email == "" || domain == "" || owner == "" {
		renderErrorPage(w, r, "400 - Bad Request: Missing email, domain, or owner parameter.")
		return
	}

	// Remove subscriber from Redis under owner-specific key
	rdb.SRem(ctx, "subs:"+owner+":"+domain, email)

	renderPage(w, r, "Unsubscription Successful", `<div class="card">
		<div class="card-body text-center">
			<h1 class="mb-4">Unsubscription Successful</h1>
			<p>You will no longer receive updates.</p>
		</div>
	</div>`) 
}
