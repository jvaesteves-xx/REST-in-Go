CREATE TABLE Invoice (
    ReferenceMonth INTEGER,
    ReferenceYear INTEGER,
    Document VARCHAR(14),
    Description VARCHAR(256),
    Amount DECIMAL(16, 2),
    IsActive BOOLEAN,
    CreatedAt  DATE,
    DeactiveAt DATE
);