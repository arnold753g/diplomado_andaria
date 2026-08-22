CREATE TABLE agencies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    department VARCHAR(30) NOT NULL DEFAULT 'Tarija',
    city VARCHAR(100) NOT NULL,
    address VARCHAR(250) NOT NULL,
    phone VARCHAR(30) NOT NULL,
    email VARCHAR(320) NOT NULL,
    manager_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    minimum_paying_age INTEGER NOT NULL DEFAULT 6 CHECK (minimum_paying_age BETWEEN 0 AND 18),
    accepts_qr BOOLEAN NOT NULL DEFAULT FALSE,
    accepts_transfer BOOLEAN NOT NULL DEFAULT FALSE,
    bank_name VARCHAR(100) NOT NULL DEFAULT '',
    account_holder VARCHAR(160) NOT NULL DEFAULT '',
    account_number VARCHAR(50) NOT NULL DEFAULT '',
    payment_instructions VARCHAR(1000) NOT NULL DEFAULT '',
    qr_image BYTEA,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (department IN ('Tarija','Chuquisaca','La Paz','Cochabamba','Oruro','Potosí','Santa Cruz','Beni','Pando')),
    CHECK (NOT accepts_qr OR octet_length(qr_image) > 0 AND qr_image IS NOT NULL),
    CHECK (NOT accepts_transfer OR (length(trim(bank_name)) > 0 AND length(trim(account_holder)) > 0 AND length(trim(account_number)) > 0))
);
CREATE INDEX agencies_name_search ON agencies (lower(name));
CREATE INDEX agencies_status_department ON agencies(status, department);
CREATE TRIGGER agencies_set_updated_at BEFORE UPDATE ON agencies FOR EACH ROW EXECUTE FUNCTION set_updated_at();
