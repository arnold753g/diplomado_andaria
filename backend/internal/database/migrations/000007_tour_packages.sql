CREATE TABLE tour_packages (
    id BIGSERIAL PRIMARY KEY,
    agency_id BIGINT NOT NULL REFERENCES agencies(id) ON DELETE RESTRICT,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_days INTEGER NOT NULL DEFAULT 1 CHECK (duration_days BETWEEN 1 AND 30),
    duration_nights INTEGER NOT NULL DEFAULT 0 CHECK (duration_nights BETWEEN 0 AND 29 AND duration_nights < duration_days),
    difficulty VARCHAR(20) NOT NULL DEFAULT '' CHECK (difficulty IN ('','easy','moderate','demanding')),
    national_price_cents INTEGER NOT NULL DEFAULT 0 CHECK (national_price_cents BETWEEN 0 AND 100000000),
    foreign_surcharge_cents INTEGER NOT NULL DEFAULT 0 CHECK (foreign_surcharge_cents BETWEEN 0 AND 100000000),
    includes JSONB NOT NULL DEFAULT '[]'::jsonb,
    excludes JSONB NOT NULL DEFAULT '[]'::jsonb,
    bring JSONB NOT NULL DEFAULT '[]'::jsonb,
    cancellation_allowed BOOLEAN NOT NULL DEFAULT FALSE,
    cancellation_notice_hours INTEGER NOT NULL DEFAULT 0 CHECK (cancellation_notice_hours BETWEEN 0 AND 8760),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, agency_id),
    CHECK (cancellation_allowed OR cancellation_notice_hours = 0)
);
CREATE INDEX tour_packages_agency ON tour_packages(agency_id, created_at DESC);
CREATE INDEX tour_packages_catalog ON tour_packages(published, created_at DESC);
CREATE TRIGGER tour_packages_set_updated_at BEFORE UPDATE ON tour_packages FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tour_package_photos (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES tour_packages(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 5),
    image BYTEA NOT NULL CHECK (octet_length(image) BETWEEN 1 AND 5242880)
);
CREATE INDEX tour_package_photos_parent ON tour_package_photos(package_id, position);

CREATE TABLE tour_package_itinerary_days (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES tour_packages(id) ON DELETE CASCADE,
    day_number INTEGER NOT NULL CHECK (day_number BETWEEN 1 AND 30),
    title VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    activities JSONB NOT NULL DEFAULT '[]'::jsonb,
    UNIQUE (package_id, day_number)
);
CREATE INDEX tour_package_itinerary_parent ON tour_package_itinerary_days(package_id, day_number);

CREATE TABLE tour_package_itinerary_attractions (
    itinerary_day_id BIGINT NOT NULL REFERENCES tour_package_itinerary_days(id) ON DELETE CASCADE,
    attraction_id BIGINT NOT NULL REFERENCES attractions(id) ON DELETE RESTRICT,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 19),
    PRIMARY KEY (itinerary_day_id, attraction_id),
    UNIQUE (itinerary_day_id, position)
);
CREATE INDEX tour_package_itinerary_attraction ON tour_package_itinerary_attractions(attraction_id);
