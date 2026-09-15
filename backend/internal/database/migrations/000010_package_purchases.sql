CREATE TABLE tour_package_purchases (
    id BIGSERIAL PRIMARY KEY,
    reference VARCHAR(40) NOT NULL UNIQUE,
    tourist_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    agency_id BIGINT NOT NULL REFERENCES agencies(id) ON DELETE RESTRICT,
    package_id BIGINT NOT NULL REFERENCES tour_packages(id) ON DELETE RESTRICT,
    departure_id BIGINT NOT NULL REFERENCES tour_package_departures(id) ON DELETE RESTRICT,
    status VARCHAR(24) NOT NULL DEFAULT 'payment_review'
        CHECK (status IN ('payment_review','confirmed','payment_rejected','cancelled','refund_pending','refunded')),
    payment_method VARCHAR(16) NOT NULL CHECK (payment_method IN ('qr','transfer')),
    national_adults INTEGER NOT NULL DEFAULT 0 CHECK (national_adults BETWEEN 0 AND 500),
    foreign_adults INTEGER NOT NULL DEFAULT 0 CHECK (foreign_adults BETWEEN 0 AND 500),
    minors JSONB NOT NULL DEFAULT '[]'::jsonb,
    free_minor_count INTEGER NOT NULL DEFAULT 0 CHECK (free_minor_count BETWEEN 0 AND 500),
    paying_minor_count INTEGER NOT NULL DEFAULT 0 CHECK (paying_minor_count BETWEEN 0 AND 500),
    capacity_count INTEGER NOT NULL CHECK (capacity_count BETWEEN 1 AND 500),
    national_unit_price_cents INTEGER NOT NULL CHECK (national_unit_price_cents >= 0),
    foreign_surcharge_cents INTEGER NOT NULL CHECK (foreign_surcharge_cents >= 0),
    total_cents INTEGER NOT NULL CHECK (total_cents > 0),
    tourist_name VARCHAR(210) NOT NULL,
    tourist_email VARCHAR(320) NOT NULL,
    tourist_phone VARCHAR(30) NOT NULL DEFAULT '',
    tourist_document VARCHAR(40) NOT NULL DEFAULT '',
    tourist_nationality VARCHAR(80) NOT NULL DEFAULT '',
    payment_proof BYTEA NOT NULL CHECK (octet_length(payment_proof) BETWEEN 1 AND 5242880),
    agency_review_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    rejection_reason TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (national_adults + foreign_adults >= 1),
    CHECK (capacity_count >= national_adults + foreign_adults),
    CHECK (
        (status = 'payment_review' AND reviewed_at IS NULL AND rejection_reason = '') OR
        (status = 'confirmed' AND reviewed_at IS NOT NULL AND rejection_reason = '') OR
        (status = 'payment_rejected' AND reviewed_at IS NOT NULL AND length(trim(rejection_reason)) >= 5) OR
        (status IN ('cancelled','refund_pending','refunded'))
    )
);

CREATE INDEX tour_package_purchases_tourist
    ON tour_package_purchases(tourist_id, created_at DESC);
CREATE INDEX tour_package_purchases_agency
    ON tour_package_purchases(agency_id, status, created_at DESC);
CREATE INDEX tour_package_purchases_departure
    ON tour_package_purchases(departure_id, status);
CREATE TRIGGER tour_package_purchases_set_updated_at
    BEFORE UPDATE ON tour_package_purchases
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
