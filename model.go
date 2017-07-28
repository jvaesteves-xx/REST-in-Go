package main

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type Invoice struct {
	ReferenceMonth int
	ReferenceYear  int
	Document       string  `json:"Document"`
	Description    string  `json:"Description"`
	Amount         float32 `json:"Amount"`
	CreatedAt      string  `json:"CreatedAt"`
	IsActive       bool
	DeactiveAt     interface{}
}

func createSelectStatement(sqlParams map[string]interface{}) (string, []interface{}) {

	sqlStatement := "SELECT * FROM invoice "
	params := make([]interface{}, 0)
	counter := 1

	// Filter: month, year, document
	if iWhere, ok := sqlParams["where"]; ok {
		where := iWhere.(map[string]string)
		sqlStatement += "WHERE "

		month, ok := where["month"]
		if ok {
			sqlStatement += "ReferenceMonth = $" + strconv.Itoa(counter) + " AND "
			params = append(params, month)
			counter++
		}

		year, ok := where["year"]
		if ok {
			sqlStatement += "ReferenceYear = $" + strconv.Itoa(counter) + " AND "
			params = append(params, year)
			counter++
		}

		document, ok := where["document"]
		if ok {
			sqlStatement += "Document = $" + strconv.Itoa(counter) + " AND "
			params = append(params, document)
			counter++
		}

		sqlStatement = sqlStatement[:len(sqlStatement)-4]
	}

	// Order by month, year, document or all of these;
	if iOrderby, ok := sqlParams["orderby"]; ok {
		sqlStatement += "ORDER BY "
		orderby := iOrderby.([]string)

		for _, ob := range orderby {
			switch ob {
			case "year":
				sqlStatement += "ReferenceYear, "
			case "month":
				sqlStatement += "ReferenceMonth, "
			case "document":
				sqlStatement += "Document, "
			}
		}

		sqlStatement = sqlStatement[:len(sqlStatement)-2] + " "
	}

	// Pagination
	limit := sqlParams["limit"].(int)
	offset := sqlParams["offset"].(int)

	sqlStatement += "LIMIT $" + strconv.Itoa(counter) + " OFFSET $" + strconv.Itoa(counter+1)
	params = append(params, limit, offset)

	return sqlStatement, params
}

func createUpdateStatement(sqlParams map[string]interface{}, month, year int, document string) (string, []interface{}) {
	var params []interface{}
	var counter int
	sqlStatement := "UPDATE invoice SET "

	for k, v := range sqlParams {
		k = strings.ToLower(k)
		counter++

		switch k {
		case "document":
			sqlStatement += "document=$" + strconv.Itoa(counter) + ", "
		case "description":
			sqlStatement += "description=$" + strconv.Itoa(counter) + ", "
			params = append(params, v)
		case "amount":
			sqlStatement += "amount=$" + strconv.Itoa(counter) + ", "
			params = append(params, v)
		case "createdat":
			// Here, it would be better if "ReferenceMonth" and "ReferenceYear" were
			// just updateable by a trigger at the database, but as it can't assumed that
			// this feature will be available at the DB, this is being done here, hardcoded
			sqlStatement += "ReferenceMonth=$" + strconv.Itoa(counter) + ", "
			sqlStatement += "ReferenceYear=$" + strconv.Itoa(counter+1) + ", "
			sqlStatement += "createdat=$" + strconv.Itoa(counter+2) + ", "
			counter += 2

			params = append(params, v.(string)[5:7], v.(string)[:4], v.(string))
		default:
			counter--
		}
	}

	sqlStatement = sqlStatement[:len(sqlStatement)-2] + " WHERE "
	sqlStatement += "ReferenceMonth = $" + strconv.Itoa(counter+1) + " AND "
	sqlStatement += "ReferenceYear = $" + strconv.Itoa(counter+2) + " AND "
	sqlStatement += "Document = $" + strconv.Itoa(counter+3)

	params = append(params, month, year, document)

	return sqlStatement, params
}

func (invoice *Invoice) CreateInvoice(db *sql.DB) error {
	month, _ := strconv.Atoi(invoice.CreatedAt[5:7])
	year, _ := strconv.Atoi(invoice.CreatedAt[:4])

	invoice.ReferenceMonth = month
	invoice.ReferenceYear = year
	invoice.IsActive = true
	invoice.DeactiveAt = nil

	_, err := db.Exec(
		`INSERT INTO invoice(ReferenceMonth, ReferenceYear, Document, Description, Amount, IsActive, CreatedAt, DeactiveAt)
		 VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		invoice.ReferenceMonth,
		invoice.ReferenceYear,
		invoice.Document,
		invoice.Description,
		invoice.Amount,
		invoice.IsActive,
		invoice.CreatedAt,
		invoice.DeactiveAt,
	)

	return err
}

func GetInvoices(db *sql.DB, params map[string]interface{}) ([]Invoice, error) {
	sqlStatement, sqlParams := createSelectStatement(params)
	rows, err := db.Query(sqlStatement, sqlParams...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	invoices := []Invoice{}

	for rows.Next() {
		var invoice Invoice
		err := rows.Scan(
			&invoice.ReferenceMonth,
			&invoice.ReferenceYear,
			&invoice.Document,
			&invoice.Description,
			&invoice.Amount,
			&invoice.IsActive,
			&invoice.CreatedAt,
			&invoice.DeactiveAt,
		)

		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (invoice *Invoice) UpdateInvoice(db *sql.DB, month, year int, document string, toUpdate map[string]interface{}) error {
	sqlStatement, params := createUpdateStatement(toUpdate, month, year, document)
	_, err := db.Exec(sqlStatement, params...)

	return err
}

func (invoice *Invoice) DeleteInvoice(db *sql.DB) error {
	today := time.Now().Format("2006-01-02")
	_, err := db.Exec(`
		UPDATE invoice
		SET isActive = false
		AND DeactiveAt = $1
		WHERE ReferenceMonth = $2
		AND ReferenceYear = $3
		AND Document = $4`,
		today,
		invoice.ReferenceMonth,
		invoice.ReferenceYear,
		invoice.Document,
	)

	return err
}
