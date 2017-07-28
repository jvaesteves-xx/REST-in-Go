package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
}

func (app *App) Initialize(user, password, dbname string) {
	err := ConnectDatabase(user, password, dbname)

	if err != nil {
		log.Fatal(err)
	}

	app.Router = mux.NewRouter()
	app.initializeRoutes()
}

func (app *App) initializeRoutes() {
	app.Router.Handle("/invoice", CreateInvoiceHandler).Methods("POST")
	app.Router.Handle("/invoices", GetInvoicesHandler).Methods("GET")
	app.Router.Handle("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}", GetInvoicesHandler).Methods("GET")
	app.Router.Handle("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}", GetInvoicesHandler).Methods("GET")
	app.Router.Handle("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", GetInvoicesHandler).Methods("GET")
	app.Router.Handle("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", UpdateInvoiceHandler).Methods("PUT")
	app.Router.Handle("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", DeleteInvoiceHandler).Methods("DELETE")
}

func (app *App) Run(port string) {
	log.Fatal(http.ListenAndServe(port, handlers.LoggingHandler(os.Stdout, app.Router)))
}
