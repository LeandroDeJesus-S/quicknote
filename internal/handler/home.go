// Package handler provides HTTP handlers.
package handler

import "net/http"

// HomeHandler handles the home page.
func HomeHandler(w http.ResponseWriter, r *http.Request) error {
	return render(
		w,
		newRenderOpts().WithPage("home.html"),
	)
}
