package main

import (
	"fmt"
	"net/http"
	"time"

	"firstAPI.jweaver11.net/internal/data" //pg 48 error
)

//Add a 'createMovieHandler' for the "Post /v1/movies" endpoint.
//Returns the plain-text placeholder response
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	//Declares an anonymous struct to thold the information that we expect to be in the HTTP request body
	//(filed naes and types are subsets of the movie struct created earlier).
	//This struct will be our *target decode destination*
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	//Initialize new json.Decoder instance to read request bodys, and then use decode method to decode the body
	//contents into the 'input' struct
	//When we call the Decode() we pass a *pointer* to the input struct as the target decode destination.
	//If there was an error during decoding, we use our generic errorResponse() helper to send 400 bad request
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

//Add a 'showMovieHandler' for the "Get /v1/movies/:id" endpoint.
//For now, retrieves the interpolated "id" parameter from current URL and include it in placeholder response
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Create a new instance of the Movie struct, containing the ID we extracted from the URL and some dummy data
	//Also notice that we deliberatelyy haven't set a value for the Year field
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	//Encode the struct to JSON and send it as the HTTP response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
