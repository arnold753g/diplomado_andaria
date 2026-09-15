CREATE TABLE tour_package_schedules (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL UNIQUE REFERENCES tour_packages(id) ON DELETE CASCADE,
    frequency_type VARCHAR(24) NOT NULL CHECK (frequency_type IN ('single','daily','specific_weekdays')),
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    weekdays JSONB NOT NULL DEFAULT '[]'::jsonb,
    departure_time TIME NOT NULL,
    meeting_time TIME NOT NULL,
    default_min_capacity INTEGER NOT NULL CHECK (default_min_capacity BETWEEN 1 AND 500),
    default_max_capacity INTEGER NOT NULL CHECK (default_max_capacity BETWEEN 1 AND 500 AND default_max_capacity >= default_min_capacity),
    booking_cutoff_hours INTEGER NOT NULL CHECK (booking_cutoff_hours BETWEEN 1 AND 720),
    maximum_advance_days INTEGER NOT NULL CHECK (maximum_advance_days BETWEEN 1 AND 365),
    default_meeting_point TEXT NOT NULL DEFAULT '',
    default_instructions TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (valid_until >= valid_from),
    CHECK (valid_until <= (valid_from + INTERVAL '1 year')::date),
    CHECK (booking_cutoff_hours < maximum_advance_days * 24)
);
CREATE INDEX tour_package_schedules_range ON tour_package_schedules(valid_from, valid_until);
CREATE TRIGGER tour_package_schedules_set_updated_at BEFORE UPDATE ON tour_package_schedules FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tour_package_departures (
    id BIGSERIAL PRIMARY KEY,
    package_id BIGINT NOT NULL REFERENCES tour_packages(id) ON DELETE CASCADE,
    schedule_id BIGINT REFERENCES tour_package_schedules(id) ON DELETE SET NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    meeting_at TIMESTAMPTZ NOT NULL,
    booking_opens_at TIMESTAMPTZ NOT NULL,
    booking_closes_at TIMESTAMPTZ NOT NULL,
    min_capacity INTEGER NOT NULL CHECK (min_capacity BETWEEN 1 AND 500),
    max_capacity INTEGER NOT NULL CHECK (max_capacity BETWEEN 1 AND 500 AND max_capacity >= min_capacity),
    held_capacity INTEGER NOT NULL DEFAULT 0 CHECK (held_capacity >= 0),
    confirmed_capacity INTEGER NOT NULL DEFAULT 0 CHECK (confirmed_capacity >= 0),
    meeting_point TEXT NOT NULL DEFAULT '',
    instructions TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('draft','open','confirmed','closed','cancelled','completed')),
    cancellation_reason TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (package_id, starts_at),
    CHECK (meeting_at <= starts_at),
    CHECK (booking_opens_at < booking_closes_at),
    CHECK (booking_closes_at < starts_at),
    CHECK (held_capacity + confirmed_capacity <= max_capacity),
    CHECK (status <> 'cancelled' OR length(trim(cancellation_reason)) > 0)
);
CREATE INDEX tour_package_departures_catalog ON tour_package_departures(package_id, status, starts_at);
CREATE INDEX tour_package_departures_schedule ON tour_package_departures(schedule_id, starts_at);
CREATE TRIGGER tour_package_departures_set_updated_at BEFORE UPDATE ON tour_package_departures FOR EACH ROW EXECUTE FUNCTION set_updated_at();
