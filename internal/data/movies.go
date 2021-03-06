package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"firstAPI.jweaver11.net/internal/validator"

	"github.com/lib/pq"
)

//Define 'MovieModel' struct which wraps a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

//Add a placeholder method for inserting a new recod in the movies table
func (m MovieModel) Insert(movie *Movie) error {
	//define the SQL query for inserting a new record in the movies table and returning the system-generated data
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	//Create an args slice containing the values for the placeholder parameters from the movie struct.
	//Declaring this slice immediately next to our SQL query helps to make it clear *What values are uses where* in query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Use the QueryRow() method to execute the SQL query on our connection pool, passing in the args slice as a variadic
	//parameter and scanning the system genereated id, created_at and version values into the movie struct
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

//Add a placeholder method for fetching a specific record fromt he movies table
func (m MovieModel) Get(id int64) (*Movie, error) {
	//The PostgreSQL bigserial type starts auto-incrementing at 1 by default, so we know non movies have an ID number less than that
	//To avoic unenecessary database call, we take a shortcut and return 'ErrRecordNotFound' error straight away
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	//Define the SQL query for retrieving the movie data
	//Makes our database sleep for 10 seconds before return response
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	//Execute the query using the QueryRow() method, passing in the providied id value as a placeholder parameter, and scan the
	//response data into the fileds of Movie struct. We convert the scan target for the genres column using the pq.Array() adapter function
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	//Handle any errors. If there was no matching movie found, Scan() will return a sql.ErrNoRows errror.
	//We check for this and return our custom 'ErrRecordNotFound' error instead
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

//Add a placeholder method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	//Declare the SQL query for updating the record and returning the new version number
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	//create an args slice containing the values for the placeholder parameters
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, //add the expected movie version
	}

	//Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Execute the SQL query. If no matching row could be found, we know th emovie version has changed (or that record
	//has been deleted) and we return our custom ErrEditConflict error
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

//Add a placeholder method for deleting a specific record from the movies table
func (m MovieModel) Delete(id int64) error {
	//Return an ErrRecordNotFound error if the movie ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	//Construct the SQL query to delete the record
	query := `
		DELETE FROM movies
		WHERE id = $1`

	//Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Execute the SQL query using the Exec() method, passing in the id variables as the value for the placeholder parameter
	//The Exec() method returns a sql.Result object
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	//Call the RowsAffected() method on teh sql.Result object to get the number of rows affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//If no rows were affected, we know that the movies table didn't contain a record with the provided ID at the moment
	//we tried to delete it. In that case we return an ErrRecordNotFound error
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

type MockMovieModel struct{}

func (m MockMovieModel) Insert(movie *Movie) error {
	//Mock the action...
	return nil
}

func (m MockMovieModel) Get(id int64) (*Movie, error) {
	//Mock the action...
	return nil, nil
}

func (m MockMovieModel) Update(movie *Movie) error {
	//Mock the action...
	return nil

}

func (m MockMovieModel) Delete(id int64) error {
	//Mock the action...
	return nil

}

type Movie struct {
	ID        int64     `json:"id"`                //Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 //Timestamp for when the movie is added to our database
	Title     string    `json:"title"`             //Movie title
	Year      int32     `json:"year,omitempty"`    //Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` //Movie runtime (in minutes)
	Genres    []string  `json:"genres,omitempty"`  //Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     `json:"version"`           //The version number starts at 1 and will be incremented each time the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")

	v.Check(validator.Unique(movie.Genres), "genres", "must not contain dupliate values")
}

//Create a new 'GetAll()' method which returns a slice of movies. We set these up to accept the various filter parameters as arguments
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	//Construct the SQL query to retrieve all movie records
	query := `
	SELECT id, created_at, title, year, runtime, genres, version
	FROM movies
	ORDER BY id`

	//Create a context with a 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Use the QueryContext() to execute the query. Returns the sql.Rows resultset with the result
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	//defer a call to rows.Close() to ensure that the resultset is closed before 'GetAll()' returns
	defer rows.Close()

	//Initialize an empty slice to hold movie data
	movies := []*Movie{}

	//use rows.Next to iterate through the rows in the resultset
	for rows.Next() {
		//Initialize an empty Movie struct to hold the data for an individual movie
		var movie Movie

		//Scan the values from row into movie struct
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		//Add the Movie struct to the slice
		movies = append(movies, &movie)
	}

	//When the rows.Next() loop has finished, call rows.Err() to retrieve any error encountered
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}
