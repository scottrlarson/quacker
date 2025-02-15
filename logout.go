package main

import (
	"net/http"
)

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Invalidate the user's cookie to log them out
	http.SetCookie(w, &http.Cookie{
		Name:   "quacker_user",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Deletes the cookie immediately
	})

	// Redirect the user to the home page
	http.Redirect(w, r, "/", http.StatusFound)
}
