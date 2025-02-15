// edit_site_handlers.go

package main

import (
	"fmt"
	"net/http"
)

func editSitePage(w http.ResponseWriter, r *http.Request) {
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

	rssURL, _ := rdb.Get(ctx, "rss:"+loggedInUser+":"+domain).Result()
	replyTo, _ := rdb.Get(ctx, "user_sites:"+loggedInUser+":"+domain).Result()

	content := fmt.Sprintf(`
		<h1 class="mb-4">Edit Site</h1>
		<form action="/editsite" method="POST">
			<input type="hidden" name="domain" value="%s">
			<div class="mb-3">
				<label for="rss" class="form-label">RSS Feed URL</label>
				<input type="url" class="form-control" id="rss" name="rss" value="%s" placeholder="Enter RSS feed URL" required>
			</div>
			<div class="mb-3">
				<label for="replyto" class="form-label">Reply-To Email</label>
				<input type="email" class="form-control" id="replyto" name="replyto" value="%s" placeholder="Enter reply-to email" required>
			</div>
			<button type="submit" class="btn btn-primary">Save Changes</button>
		</form>`, domain, rssURL, replyTo)

	renderPage(w, r, "Edit Site", content)
}

func editSite(w http.ResponseWriter, r *http.Request) {
	domain := r.FormValue("domain")
	rss := r.FormValue("rss")
	replyTo := r.FormValue("replyto")

	if domain == "" || rss == "" || replyTo == "" {
		renderErrorPage(w, r, "400 - Bad Request: Missing parameters.")
		return
	}

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

	cookie, err := r.Cookie("quacker_user")
	if err != nil {
		renderErrorPage(w, r, "401 - Unauthorized: Please log in.")
		return
	}
	loggedInUser := cookie.Value

	domainKey := "user_sites:" + loggedInUser + ":" + domain
	rdb.Set(ctx, domainKey, replyTo, 0)
	rdb.Set(ctx, "rss:"+loggedInUser+":"+domain, rss, 0)
	http.Redirect(w, r, "/sites?success=Site+updated", http.StatusFound)
}
