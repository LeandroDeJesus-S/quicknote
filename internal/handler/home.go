package handler

import "net/http"

func HomeHandler(w http.ResponseWriter, r *http.Request) error {
	return render(
		w,
		newRenderOpts().WithPage("home.html"),
	)
}
