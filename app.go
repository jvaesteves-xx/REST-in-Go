package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"time"

	"io/ioutil"

	"bytes"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// ### Utils ###
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// ### --- ###

func (app *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	app.DB, err = sql.Open("postgres", connectionString)

	if err != nil {
		log.Fatal(err)
	}

	app.Router = mux.NewRouter()
	app.initializeRoutes()
}

func (app *App) Run(port string) {
	log.Fatal(http.ListenAndServe(port, app.Router))
}

func (app *App) initializeRoutes() {
	app.Router.HandleFunc("/invoice", app.createInvoice).Methods("POST")
	app.Router.HandleFunc("/invoices", app.getInvoices).Methods("GET")
	app.Router.HandleFunc("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}", app.getInvoices).Methods("GET")
	app.Router.HandleFunc("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}", app.getInvoices).Methods("GET")
	app.Router.HandleFunc("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", app.getInvoices).Methods("GET")
	app.Router.HandleFunc("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", app.updateInvoice).Methods("PUT")
	app.Router.HandleFunc("/invoices/{year:19[5-9][0-9]|20[0-9]{2}}/{month:[1-9]|1[0-2]}/{document:[a-zA-Z0-9]{14}}", app.deleteInvoice).Methods("DELETE")
}

func (app *App) getInvoices(response http.ResponseWriter, request *http.Request) {
	sqlParams, where := make(map[string]interface{}), mux.Vars(request)
	limit, err := strconv.Atoi(request.FormValue("per_page"))

	if err != nil || limit > 400 || limit < 1 {
		limit = 100
	}
	sqlParams["limit"] = limit

	offset, err := strconv.Atoi(request.FormValue("page"))
	if err != nil || offset < 0 {
		offset = 0
	}
	sqlParams["offset"] = offset * limit

	orderby := request.URL.Query()["order"]

	if len(where) > 0 {
		sqlParams["where"] = where
	}

	if len(orderby) > 0 {
		sqlParams["orderby"] = orderby
	}

	invoices, err := getInvoices(app.DB, sqlParams)
	if err != nil {
		respondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(response, http.StatusOK, invoices)
}

func (app *App) createInvoice(response http.ResponseWriter, request *http.Request) {
	var invoice Invoice
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&invoice); err != nil {
		respondWithError(response, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer request.Body.Close()

	today := time.Now().Unix()
	createdAt, err := time.Parse("2006-01-02", invoice.CreatedAt)

	if err != nil || len(invoice.Document) != 14 || len(invoice.Description) > 256 || createdAt.Unix() > today {
		respondWithError(response, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := invoice.createInvoice(app.DB); err != nil {
		respondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(response, http.StatusCreated, invoice)
}

func (app *App) updateInvoice(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)

	year, errY := strconv.Atoi(vars["year"])
	month, errM := strconv.Atoi(vars["month"])
	document := vars["document"]

	if errY != nil || errM != nil {
		respondWithError(response, http.StatusBadRequest, "Invalid product year/month/document ID")
		return
	}

	var fieldsToUpdate map[string]interface{}
	var invoice Invoice
	data, errIO := ioutil.ReadAll(request.Body)
	errJ := json.Unmarshal(data, &fieldsToUpdate)
	decoder := json.NewDecoder(bytes.NewReader(data))

	if errD := decoder.Decode(&invoice); errD != nil || errIO != nil || errJ != nil {
		respondWithError(response, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer request.Body.Close()

	if err := invoice.updateInvoice(app.DB, int16(month), int16(year), document, fieldsToUpdate); err != nil {
		respondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(response, http.StatusOK, map[string]string{"result": "success"})
}

func (app *App) deleteInvoice(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)

	year, errY := strconv.Atoi(vars["year"])
	month, errM := strconv.Atoi(vars["month"])
	document := vars["document"]

	if errY != nil || errM != nil {
		respondWithError(response, http.StatusBadRequest, "Invalid product year/month/document ID")
		return
	}

	invoice := Invoice{ReferenceMonth: int16(month), ReferenceYear: int16(year), Document: document}
	if err := invoice.deleteInvoice(app.DB); err != nil {
		respondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(response, http.StatusOK, map[string]string{"result": "success"})
}
