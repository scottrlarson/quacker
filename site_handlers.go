// site_handlers.go

package main

import (
	"encoding/xml"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var emailReg = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func addSitePage(w http.ResponseWriter, r *http.Request) {
	content := `
		<h1 class="mb-4">Add a New Site</h1>
		<form action="/addsite" method="POST">
			<div class="mb-3">
				<label for="rss" class="form-label">RSS Feed URL</label>
				<input type="url" class="form-control" id="rss" name="rss" placeholder="Enter RSS feed URL" required>
			</div>
			<div class="mb-3">
				<label for="replyto" class="form-label">Reply-To Email</label>
				<input type="email" class="form-control" id="replyto" name="replyto" placeholder="Enter reply-to email" required>
			</div>
			<button type="submit" class="btn btn-primary">Save Site</button>
		</form>
	`
	renderPage(w, r, "Add a New Site", content)
}

func addSite(w http.ResponseWriter, r *http.Request) {
	rss := r.FormValue("rss")
	replyTo := r.FormValue("replyto")

	if !emailReg.MatchString(replyTo) {
		http.Redirect(w, r, "/sites?error=Invalid+email+address", http.StatusFound)
		return
	}

	resp, err := http.Get(rss)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Redirect(w, r, "/sites?error=Invalid+RSS+URL", http.StatusFound)
		return
	}
	defer resp.Body.Close()

	rssData := RSS{}
	if err := xml.NewDecoder(resp.Body).Decode(&rssData); err != nil || rssData.XMLName.Local != "rss" || len(rssData.Channel.Items) == 0 {
		http.Redirect(w, r, "/sites?error=Invalid+RSS+format", http.StatusFound)
		return
	}

	domain := strings.Split(rss, "/")[2]
	cookie, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	loggedInUser := cookie.Value

	rdb.Set(ctx, "user_sites:"+loggedInUser+":"+domain, replyTo, 0)
	rdb.Set(ctx, "site_created:"+loggedInUser+":"+domain, time.Now().Format(time.RFC3339), 0)

	http.Redirect(w, r, "/sites?success=Site+added", http.StatusFound)
}

func deleteSite(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		renderErrorPage(w, r, "400 - Bad Request: Missing domain parameter.")
		return
	}
	cookie, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	loggedInUser := cookie.Value

	// Delete site entry
	rdb.Del(ctx, "user_sites:"+loggedInUser+":"+domain)

	// Delete all subscribers for the site
	rdb.Del(ctx, "subs:"+loggedInUser+":"+domain)

	// Delete all sent post records for the site
	sentKeys, _ := rdb.Keys(ctx, "sent:"+loggedInUser+":"+domain+":*").Result()
	for _, key := range sentKeys {
		rdb.Del(ctx, key)
	}

	// Delete all stored RSS posts for the site
	rssKeys, _ := rdb.Keys(ctx, "rss:"+loggedInUser+":"+domain+":*").Result()
	for _, key := range rssKeys {
		rdb.Del(ctx, key)
	}

	http.Redirect(w, r, "/sites?success=Site+deleted", http.StatusFound)
}
