package main

import (
	"errors"
	"fmt"
	"net/http"

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

	//Call the Get() method to fetch the data for a specific movie. We also need to use the Errors.Is() function
	//to check if it returns a data.ErrRecordNotFound error, in which case we send a 404 Not Found response to the client
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Encode the struct to JSON and send it as the HTTP response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	//Extradct the movie ID from the URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Fetch the existing movie record from the database, esnding a 404 Not Found response to the client if we cant find matching record
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Declare an input struct to hold the expected fata from client
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	//Read the JSON request body data into the iput strucdt
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//If the input.Title value is nil then we know no corresponding "title" key value pair was provided in the JSON body request.
	//We leave the movie record unchanged, otherwise we update the movie record with new title.
	//Since it is a pointer now, we need to dereference the pointer using the * operator to get underlying value before
	//assigning it to movie record
	if input.Title != nil {
		movie.Title = *input.Title
	}

	//We also do the same for other fileds in input struct
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres //We dont deference a slice
	}

	//Validate the updated movie record, sending the client 422 uprocessable entit response if any checks fail
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	//Intercept any ErrEditConflict error to call the new editConflictResponse() helper
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Write thee update movie record in a JSON responsee
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	//Extract the movie ID from the URl
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	//Delte the movie from the database, sending a 404 Not Found response to the client if there isn't a matching record
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	//Return a 200 OK status code along with a success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
