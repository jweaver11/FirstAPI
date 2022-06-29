package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"firstAPI.jweaver11.net/internal/data"
	//import pq driver so that it can register itself with the database/sql package.
	_ "github.com/lib/pq" //Uses black identifier so compiler doesn't complain its not being used.
)

//Declares var 'version' as a global constant string containing the application version number. Will be dynamic later.
const version = "1.0.0"

//Defines 'config' as a struct to hold all configuration settings for our app.
type config struct {
	port int    //'port' is the network port for the server to listen on
	env  string //'env' is the name of current operating environment for the app
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

//Declares 'application' as a struct to hold dependecies for our HTTP handlers, helpers, and middleware. Will grow as we build
type application struct {
	config config      //copy of config struct
	logger *log.Logger //'logger' is a logger
	models data.Models
}

//MAIN FUNCTION***************************************************************************************************************
func main() {
	var cfg config //Declares 'cfg' as an instance of the config struct

	//Reads the value of the port and env command-line flags into the config struct. We set default port number to '4000'
	//and the environment to 'development' if no other flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	//Read the DSN value from the db-dsn command-line flag into the config struct
	//Default to using our development DSN if no flag is provided
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("FIRSTAPIDB_DB_DSN"), "PostgreSQL DSN") //Needs person change

	//Read connection pool settings from command-line flags into config struct.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	//Initialize 'logger' a a new logger to write messages to the standard out stream
	//Previxed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Printf("database connection pool established")

	//Declares 'app' as an instance of application struct, containing the config struct and the logger
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	//Declares a HTTP server with some sensible timeout settings, which listens to provided port in the config struct
	//uses the 'routes.go' as the server handler
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	//Starts the HTTP server.
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)

	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	//use the sqp.Open() to create an empty connection pool, using the DSN from the config strucg
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	//Set max number of open connections in the pool
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	//Set max number of idle connections in the pool.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	//Use the 'time.ParseDuration()' function to conver the idle timeout duration string to a 'time.Duration' type
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	//Set max idle timeout
	db.SetConnMaxIdleTime(duration)

	//create context with a 5 second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//use 'PingContext()' to establish a new connection to the database, passing in the context we created above
	//as the parameter. If the connection couldn't be established successfully within the 6 second deadline,
	//will return an error
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//Page 204 - chapter 9.3 doesnt run correctly
//command to run - git bash - "go run ./cmd/api"
