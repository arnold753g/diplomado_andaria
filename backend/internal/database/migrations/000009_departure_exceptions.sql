ALTER TABLE tour_package_departures
    ADD COLUMN is_exception BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN modified_by_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN cancelled_at TIMESTAMPTZ;

ALTER TABLE tour_package_departures
    ADD CONSTRAINT tour_package_departures_cancelled_at_check
    CHECK (status <> 'cancelled' OR cancelled_at IS NOT NULL);

CREATE INDEX tour_package_departures_exceptions
    ON tour_package_departures(package_id, is_exception, starts_at);
