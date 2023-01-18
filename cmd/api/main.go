package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = 8080

type application struct {
	Domain       string
	Dsn          string
	Db           repository.DatabaseRepo
	Auth         auth
	JwtSecret    string
	JwtIssuer    string
	JwtAudience  string
	CookieDomain string
}

func main() {
	// set application config
	var app application

	// read flags from command line
	flag.StringVar(&app.Dsn, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.StringVar(&app.JwtSecret, "jwt-secret", "development-secret", "signing secret")
	flag.StringVar(&app.JwtIssuer, "jwt-issuer", "example.com", "signing issuer")
	flag.StringVar(&app.JwtAudience, "jwt-audience", "example.com", "signing audience")
	flag.StringVar(&app.Domain, "domain", "example.com", "domain")
	flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie domain")
	
	flag.Parse()

	// connect to database
	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	app.Db = &dbrepo.PostgresDbRepo{Db: conn}
	defer app.Db.Connection().Close()

	app.Auth = auth{
		Issuer: app.JwtIssuer,
		Audience: app.JwtAudience,
		Secret: app.JwtSecret,
		TokenExpiry: time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath: "/",
		CookieName: "_unsecure_Host-refresh_token",
		CookieDomain: app.CookieDomain,
	}

	// start web server
	log.Println("Starting application on port", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
