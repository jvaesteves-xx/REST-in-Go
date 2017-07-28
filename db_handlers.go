package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var dbConnection *sql.DB

func ConnectDatabase(username, password, dbname string) error {

	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", username, password, dbname)

	var err error
	dbConnection, err = sql.Open("postgres", connectionString)

	return err
}

var CreateInvoiceHandler = AuthMiddleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

	var invoice Invoice
	decoder := json.NewDecoder(request.Body)

	if err := decoder.Decode(&invoice); err != nil {
		RespondWithError(response, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer request.Body.Close()

	today := time.Now().Unix()
	createdAt, err := time.Parse("2006-01-02", invoice.CreatedAt)

	if err != nil || len(invoice.Document) != 14 || len(invoice.Description) > 256 || createdAt.Unix() > today {
		RespondWithError(response, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := invoice.CreateInvoice(dbConnection); err != nil {
		RespondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(response, http.StatusCreated, invoice)
}))

var GetInvoicesHandler = AuthMiddleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

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

	invoices, err := GetInvoices(dbConnection, sqlParams)
	if err != nil {
		RespondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(response, http.StatusOK, invoices)
}))

var UpdateInvoiceHandler = AuthMiddleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)

	year, errY := strconv.Atoi(vars["year"])
	month, errM := strconv.Atoi(vars["month"])
	document := vars["document"]

	if errY != nil || errM != nil {
		RespondWithError(response, http.StatusBadRequest, "Invalid product year/month/document ID")
		return
	}

	var fieldsToUpdate map[string]interface{}
	var invoice Invoice
	data, errIO := ioutil.ReadAll(request.Body)
	errJ := json.Unmarshal(data, &fieldsToUpdate)
	decoder := json.NewDecoder(bytes.NewReader(data))

	if errD := decoder.Decode(&invoice); errD != nil || errIO != nil || errJ != nil {
		RespondWithError(response, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer request.Body.Close()

	if err := invoice.UpdateInvoice(dbConnection, month, year, document, fieldsToUpdate); err != nil {
		RespondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(response, http.StatusOK, map[string]string{"result": "success"})
}))

var DeleteInvoiceHandler = AuthMiddleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)

	year, errY := strconv.Atoi(vars["year"])
	month, errM := strconv.Atoi(vars["month"])
	document := vars["document"]

	if errY != nil || errM != nil {
		RespondWithError(response, http.StatusBadRequest, "Invalid product year/month/document ID")
		return
	}

	invoice := Invoice{ReferenceMonth: month, ReferenceYear: year, Document: document}
	if err := invoice.DeleteInvoice(dbConnection); err != nil {
		RespondWithError(response, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJSON(response, http.StatusOK, map[string]string{"result": "success"})
}))
