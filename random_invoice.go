package main

import (
	"math/rand"
	"strconv"
	"time"
)

func generateRandomString(length int) string {

	var randomString string

	for i := 0; i < length; i++ {
		randomString += string(rand.Intn(26) + 65)
	}

	return randomString
}

func generateRandomDate() string {

	randomDate := time.Unix(0, rand.Int63n(time.Now().UnixNano()))
	return randomDate.Format("2006-01-02")
}

func GenerateRandomInvoice() Invoice {

	var invoice Invoice

	invoice.CreatedAt = generateRandomDate()
	invoice.ReferenceMonth, _ = strconv.Atoi(invoice.CreatedAt[5:7])
	invoice.ReferenceYear, _ = strconv.Atoi(invoice.CreatedAt[:4])
	invoice.Document = generateRandomString(14)
	invoice.IsActive = true
	invoice.DeactiveAt = nil
	invoice.Amount = float32(rand.Int63n(1e6))
	invoice.Description = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nisi nibh, molestie a euismod non, feugiat et justo. Nullam id diam est. Vivamus gravida eget arcu a bibendum. Duis venenatis tellus ut turpis posuere maximus. Vivamus malesuada cras amet."

	return invoice
}
