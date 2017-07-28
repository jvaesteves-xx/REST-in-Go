package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

var app App
var apiToken string

func ensureTableExists(conn *sql.DB) error {
	const tableCreationQuery = `CREATE TABLE IF NOT EXISTS invoice(
		ReferenceMonth INTEGER,
		ReferenceYear INTEGER,
		Document VARCHAR(14),
		Description VARCHAR(256),
		Amount DECIMAL(16, 2),
		IsActive BOOLEAN,
		CreatedAt  DATE,
		DeactiveAt DATE
    )`

	_, err := conn.Exec(tableCreationQuery)
	return err
}

func prepareDatabase(username, password, dbname string) error {

	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", username, password, dbname)

	conn, err := sql.Open("postgres", connectionString)

	if err != nil {
		return err
	}

	if err := ensureTableExists(conn); err != nil {
		return err
	}

	// Clear table
	_, err = conn.Exec("DELETE FROM invoice")

	// Populate table. OBS.: Worst way!
	for i := 0; i < 404; i++ {
		invoice := GenerateRandomInvoice()
		err = invoice.CreateInvoice(conn)
	}

	return err
}

func executeRequest(request *http.Request, apiToken string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	request.Header.Add("Authorization", apiToken)
	app.Router.ServeHTTP(response, request)

	return response
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func validateInvoice(t *testing.T, body io.Reader) Invoice {
	var invoice Invoice

	data, errIO := ioutil.ReadAll(body)
	decoder := json.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(&invoice); errIO != nil || err != nil {
		t.Errorf("Invalid Invoice\n")
	}

	return invoice
}

func checkInvoiceValues(t *testing.T, invoice, correctInvoice Invoice) {

	if !(invoice.ReferenceMonth == correctInvoice.ReferenceMonth &&
		invoice.ReferenceYear == correctInvoice.ReferenceYear &&
		invoice.Document == correctInvoice.Document &&
		invoice.Description == correctInvoice.Description &&
		invoice.Amount == correctInvoice.Amount &&
		invoice.IsActive == correctInvoice.IsActive &&
		invoice.CreatedAt == correctInvoice.CreatedAt &&
		invoice.DeactiveAt == correctInvoice.DeactiveAt) {
		t.Errorf("One of the Invoice values is incorrect:\n")
		t.Error(invoice)
		t.Error(correctInvoice)
	}
}

func getAPIKey() (string, error) {
	url := os.Getenv("API_ISSUER") + "oauth/token"

	payload := strings.NewReader(`{
		"client_id":"` + os.Getenv("CLIENT_ID") + `",
		"client_secret":"` + os.Getenv("CLIENT_SECRET") + `",
		"audience":"` + os.Getenv("API_AUDIENCE") + `",
		"grant_type":"client_credentials"
	}`)

	request, _ := http.NewRequest("POST", url, payload)

	request.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(request)

	if err == nil && res.StatusCode == http.StatusOK {
		defer res.Body.Close()
		data, _ := ioutil.ReadAll(res.Body)

		var resJSON map[string]interface{}
		json.Unmarshal(data, &resJSON)

		return resJSON["token_type"].(string) + " " + resJSON["access_token"].(string), nil
	}

	return "", err
}

func insertInvoice(t *testing.T, invoiceJSON string) *httptest.ResponseRecorder {
	payload := []byte(invoiceJSON)

	request, _ := http.NewRequest("POST", "/invoice", bytes.NewBuffer(payload))
	response := executeRequest(request, apiToken)

	checkResponseCode(t, http.StatusCreated, response.Code)

	return response
}

func getInvoicesLength(t *testing.T, body []byte) int {
	var invoices []map[string]interface{}

	errU := json.Unmarshal(body, &invoices)

	if errU != nil {
		t.Errorf("Invalid response!\n")
	}

	return len(invoices)
}

func perPageTests(t *testing.T, perPage, realLimit int) {
	request, _ := http.NewRequest("GET", "/invoices?per_page="+strconv.Itoa(perPage), nil)
	response := executeRequest(request, apiToken)

	checkResponseCode(t, http.StatusOK, response.Code)

	body, _ := ioutil.ReadAll(response.Body)

	if length := getInvoicesLength(t, body); length != realLimit {
		t.Errorf("Amount of returned invoices incorrect: JSON(%d) != REAL(%d)!\n", length, realLimit)
	}
}

func filterTests(t *testing.T, path string) {
	request, _ := http.NewRequest("GET", path, nil)
	response := executeRequest(request, apiToken)

	checkResponseCode(t, http.StatusOK, response.Code)

	body, _ := ioutil.ReadAll(response.Body)

	if length := getInvoicesLength(t, body); length < 1 {
		t.Errorf("The filter did not return any result [%d]\n", length)
	}
}

func TestMain(m *testing.M) {

	username := os.Getenv("APP_DB_USERNAME_SANDBOX")
	password := os.Getenv("APP_DB_PASSWORD_SANDBOX")
	dbname := os.Getenv("APP_DB_NAME_SANDBOX")

	prepareDatabase(username, password, dbname)
	app = App{}
	app.Initialize(username, password, dbname)
	apiToken, _ = getAPIKey()

	code := m.Run()

	os.Exit(code)
}

func TestUnauthorizedAccess(t *testing.T) {

	request, _ := http.NewRequest("GET", "/invoices", nil)
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)

	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestCreateInvoice(t *testing.T) {

	response := insertInvoice(t, `{
		"ReferenceMonth": 2,
		"ReferenceYear": 2006,
		"Document": "ABCDEFGHIJKLMN",
		"Description": "Lorem Ipsum",
		"Amount": 123.45,
		"IsActive": false,
		"CreatedAt": "2017-06-19",
		"DeactiveAt": null,
		"Lorem": "ipsum"
	}`)

	invoice := validateInvoice(t, response.Body)
	checkInvoiceValues(t, invoice, Invoice{
		ReferenceMonth: 6,
		ReferenceYear:  2017,
		Document:       "ABCDEFGHIJKLMN",
		Description:    "Lorem Ipsum",
		Amount:         123.45,
		IsActive:       true,
		CreatedAt:      "2017-06-19",
		DeactiveAt:     nil,
	})
}

func TestInvoicePagination(t *testing.T) {
	perPageTests(t, 100, 100)
	perPageTests(t, 999, 100)
	perPageTests(t, 400, 400)
	perPageTests(t, 401, 100)
	perPageTests(t, 1, 1)
	perPageTests(t, 0, 100)
}

func TestInvoiceFilter(t *testing.T) {
	insertInvoice(t, `{
		"Document": "12345678901234",
		"Description": "Ipsum Lorem",
		"Amount": 543.21,
		"CreatedAt": "2013-09-05"
	}`)

	filterTests(t, "/invoices/2013")
	filterTests(t, "/invoices/2013/9")
	filterTests(t, "/invoices/2013/9/12345678901234")
}

func TestUpdateInvoice(t *testing.T) {
	insertInvoice(t, `{
		"Document": "43210987654321",
		"Description": "losum iprem",
		"Amount": 345.21,
		"CreatedAt": "2015-09-05"
	}`)

	payload := []byte(`{
		"Amount": 612.45,
		"IsActive": false,
		"CreatedAt": "2016-05-01"
	}`)

	request, _ := http.NewRequest("PUT", "/invoices/2015/9/43210987654321", bytes.NewBuffer(payload))
	response := executeRequest(request, apiToken)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestDeleteInvoice(t *testing.T) {
	insertInvoice(t, `{
		"Document": "43210ABCD54321",
		"Description": "mussum iprem",
		"Amount": 999.21,
		"CreatedAt": "2015-09-05"
	}`)

	payload := []byte(`{
		"Amount": 612.45,
		"IsActive": false,
		"CreatedAt": "2016-05-01"
	}`)

	request, _ := http.NewRequest("DELETE", "/invoices/2015/9/43210ABCD54321", bytes.NewBuffer(payload))
	response := executeRequest(request, apiToken)

	checkResponseCode(t, http.StatusOK, response.Code)
}
