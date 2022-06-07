package main

import (
	"fmt"
	"net/http"
)

//Add a 'createMovieHandler' for the "Post /v1/movies" endpoint.
//Returns the plain-text placeholder response
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

//Add a 'showMovieHandler' for the "Get /v1/movies/:id" endpoint.
//For now, retrieves the interpolated "id" parameter from current URL and include it in placeholder response
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	//Otherwise, interpolate the movie ID in a placeholder response
	fmt.Fprintf(w, "show the details of the movie %d\n", id)
}
