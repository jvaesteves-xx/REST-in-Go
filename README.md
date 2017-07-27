# REST-in-Go

O desafio que propomos é escrever uma API REST, usando [Go](https://golang.org/) e podendo usar qualquer framework que desejar. 
Nesta teremos um resource Invoice (Nota Fiscal). Abaixo, algumas especificações:

## Métodos a serem implementados
Implementar métodos:
  1. GET
  2. POST
  4. PUT
  3. DELETE

Atenção:
Utilizar os status codes do HTTP corretamente de acordo com cada operação.

O método DELETE não executa uma deleção física, e sim uma deleção lógica.

Na listagem é necessário ter:
  1. Paginação;
  2. Filtro por mês, ano ou documento;
  3. Escolher a ordenação por mês, ano, documento ou combinações entre;

## Autenticação
Temos que autenticar através de token de aplicação;

## Persistência
Utilizar um Banco de Dados, a sua escolha, poder ser SQLite, MariaDB, etc;

Não utilizar ORMs, fazer a queries sem o auxílio de Hibernate, SQLAlchemy, etc;

## Domínio / Modelo
    Invoice
        ReferenceMonth : INTEGER
        ReferenceYear : INTEGER
        Document : VARCHAR(14)
        Description : VARCHAR(256)
        Amount : DECIMAL(16, 2)
        IsActive : TINYINT
        CreatedAt  : DATETIME
        DeactiveAt : DATETIME

O código deve estar no Github ou BitBucket com o repositório público. Responda esse email com o link do repositório. 

O prazo fica a seu critério, não importa quanto tempo, mas o importante é que você cumpra os seus prazos.

O ideal é que a aplicação cumpra todos os requisitos, mas se isso não for possível, gostaríamos de ver o seu código mesmo assim.