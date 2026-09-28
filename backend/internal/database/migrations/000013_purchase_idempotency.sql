-- Preserve accepted request keys so delayed mobile retries cannot buy twice.
CREATE TABLE purchase_requests (
    tourist_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    request_key VARCHAR(128) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    purchase_id BIGINT NOT NULL REFERENCES tour_package_purchases(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tourist_id, request_key)
);
