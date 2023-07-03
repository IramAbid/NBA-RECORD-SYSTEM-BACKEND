package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rashidalam9678/nba_management_system/internal/database"
	"github.com/rashidalam9678/nba_management_system/internal/driver"
)

type config struct{
	port int
}

type application struct{
	config config
	infoLog *log.Logger
	errorLog *log.Logger
	models database.Models
	environment string

}

func main(){
	var cfg config
	cfg.port=8081

	infoLog:= log.New(os.Stdout,"INFO\t",log.Ldate|log.Ltime)
	errorLog:= log.New(os.Stdout,"INFO\t",log.Ldate|log.Ltime|log.Lshortfile)

	dsn := os.Getenv("DSN")
	environment:=os.Getenv("ENV")

	db,err:= driver.ConnectPostgres(dsn)
	if err!= nil{
		log.Fatal("Can not connect to database")
	}
	defer db.SQL.Close()

	app := &application{
		config: cfg,
		infoLog: infoLog,
		errorLog: errorLog,
		models: database.New(db.SQL),
		environment: environment,
	}

	err= app.Serve()
	if err!=nil{
		log.Fatal(err)
	}



}

func (app *application) Serve() error{
	app.infoLog.Println("API Listening on Port : ",app.config.port)
	srv:=&http.Server{
		Addr: fmt.Sprintf(":%d",app.config.port),
		Handler: app.routes(),
	}

	return srv.ListenAndServe()
}