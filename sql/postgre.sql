CREATE TABLE bank_accounts
(
    id      SERIAL
        CONSTRAINT bank_accounts_pk
            PRIMARY KEY,
    name    VARCHAR(35)   NOT NULL,
    surname VARCHAR(35)   NOT NULL,
    balance INT DEFAULT 0 NOT NULL
);

INSERT INTO bank_accounts (name, surname, balance)
VALUES ('Johny', 'Depp', 1500),
       ('Bred', 'Pit', 3000),
       ('Angelina', 'Jolie', 5000);
