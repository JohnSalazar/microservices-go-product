-- DROP DATABASE IF EXISTS products;

-- CREATE DATABASE products
--     WITH
--     OWNER = postgres
--     ENCODING = 'UTF8'
--     LC_COLLATE = 'en_US.utf8'
--     LC_CTYPE = 'en_US.utf8'
--     TABLESPACE = pg_default
--     CONNECTION LIMIT = -1
--     IS_TEMPLATE = False;

GRANT ALL ON DATABASE products TO PUBLIC;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres
GRANT ALL ON TABLES TO PUBLIC;

DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS stores CASCADE;

CREATE TABLE products
(
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(500) NOT NULL CHECK ( name <> '' ),
    slug VARCHAR(600) NOT NULL CHECK ( slug <> '' ),
    description VARCHAR(10000),
    price numeric(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    version integer NOT NULL DEFAULT 0,
    deleted boolean NOT NULL DEFAULT false,
    image VARCHAR(500)
);

CREATE TABLE stores
(
    id UUID PRIMARY KEY NOT NULL,
    "productid" UUID NOT NULL REFERENCES products(id),
    booked_at TIMESTAMP WITH TIME ZONE,
    sold boolean NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    version integer NOT NULL DEFAULT 0,
    deleted boolean NOT NULL DEFAULT false,
    CONSTRAINT fk_product FOREIGN KEY ("productid") REFERENCES products (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);