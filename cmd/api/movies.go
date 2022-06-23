package main

import (
	"fmt"
	"net/http"
	"time"

	"firstAPI.jweaver11.net/internal/data"
	"firstAPI.jweaver11.net/internal/validator"
)

//Add a 'createMovieHandler' for the "Post /v1/movies" endpoint.
//Returns the plain-text placeholder response
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	//Declares an anonymous struct to thold the information that we expect to be in the HTTP request body
	//(filed naes and types are subsets of the movie struct created earlier).
	//This struct will be our *target decode destination*
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
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

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Call the Insert() method on our movies model, passing in a pointer to the validated movie struct. This will
	//create a record in the database and update the movie struct with the system-generated information
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//When sending a HTTP response, we want to include a location header to let the client know which URL
	//they can find the newly-created resource at. We make an empty http.Header map and then use the Set()
	//method to add a new Loacation header, interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	//Write a JSON response with a 201 Create status code, the movie data in the response body, and Location header
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
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
