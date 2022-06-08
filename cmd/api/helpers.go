package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{} //Define an envelope type

//Retrieve the "id" URL parameter from the current request context, then convert it to
//an integer and return it. If the operation isn't successful, return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

//Define a 'writeJSON' helper for sending responses. Takes the destination http.ResponseWriter, the HTTP status code to send, the data to encode in JSON,
//and a header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	//Use the json.MarshalIndent() function so that the whitespace is added to the encoded Json
	//Here we use no line prefix ("") and tab indents ("\t") for each element
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	//Append a newline to make it easier to view in terminal applications
	js = append(js, '\n')

	//At this point we know there won't be any more errors before writing, so its safe to add headers.
	//We loop through header map and add each header to the http.ResponseWriter header map
	//Its ok if the head map provided is nil. Won't throw an error.
	for key, value := range headers {
		w.Header()[key] = value
	}

	//Add the "Content-Type: application/json" header, then write the status code and JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}