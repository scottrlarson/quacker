// download_subscribers_handler.go

package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
)

func downloadSubscribers(w http.ResponseWriter, r *http.Request) {
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

	// Retrieve the list of subscribers for the logged-in user's site from Redis
	subscribers, err := rdb.SMembers(ctx, "subs:"+loggedInUser+":"+domain).Result()
	if err != nil {
		renderErrorPage(w, r, "500 - Internal Server Error: Unable to fetch subscribers.")
		return
	}

	// Set headers for CSV file download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_subscribers.csv", domain))

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header
	csvWriter.Write([]string{"Email Address"})

	// Write subscriber data
	for _, email := range subscribers {
		csvWriter.Write([]string{email})
	}
}
