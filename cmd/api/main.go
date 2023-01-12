package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
)

const port = 8080

type application struct {
	Domain string
	Dsn string
	Db *sql.DB
}

func main() {
	// set application config
	var app application

	// read flags from command line
	flag.StringVar(&app.Dsn, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.Parse()

	// connect to database
	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	app.Db = conn
	defer app.Db.Close()

	app.Domain = "example.com"

	// start web server
	log.Println("Starting application on port", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}