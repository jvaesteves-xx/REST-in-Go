# REST-in-Go

This repository is part of the Stone Challenge that proposed the development of a REST API using [Go](https://golang.org/) and any other framework from the language.

The API would serve the resource Invoice, using the HTTP methods, correctly to their specification:

  1. GET
  2. POST
  4. PUT
  3. DELETE

**OBS.:** The DELETE method will not delete the phisical object, just logically.

The requirements of the challenge included:

  1. Pagination;
  2. Filter by month, year or document;
  3. Order by any combination of year, month and document;
  4. Authenticate the application via token;
  5. Make queries without ORMs or Hibernate, SQLAlchemy etc;

## Model
    Invoice
        ReferenceMonth : INTEGER
        ReferenceYear : INTEGER
        Document : VARCHAR(14)
        Description : VARCHAR(256)
        Amount : DECIMAL(16, 2)
        IsActive : TINYINT
        CreatedAt  : DATETIME
        DeactiveAt : DATETIME