package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

//Declares var 'version' as a global constant string containing the application version number. Will be dynamic later.
const version = "1.0.0"

//Defines 'config' as a struct to hold all configuration settings for our app.
type config struct {
	port int    //'port' is the network port for the server to listen on
	env  string //'env' is the name of current operating environment for the app
}

//Declares 'application' as a struct to hold dependecies for our HTTP handlers, helpers, and middleware. Will grow as we build
type application struct {
	config config      //copy of config struct
	logger *log.Logger //'logger' is a logger
}

//MAIN FUNCTION***************************************************************************************************************
func main() {
	var cfg config //Declares 'cfg' as an instance of the config struct

	//Reads the value of the port and env command-line flags into the config struct. We set default port number to '4000'
	//and the environment to 'development' if no other flags are provided
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	//Initialize 'logger' a a new logger to write messages to the standard out stream
	//Previxed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	//Declares 'app' as an instance of application struct, containing the config struct and the logger
	app := &application{
		config: cfg,
		logger: logger,
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
	err := srv.ListenAndServe()
	logger.Fatal(err)

	//Page 31
}
