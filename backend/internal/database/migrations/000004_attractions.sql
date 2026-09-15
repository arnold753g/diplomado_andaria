CREATE TABLE attractions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(30) NOT NULL CHECK (category IN ('Naturaleza','Cultura e historia','Aventura','Gastronomía','Enoturismo','Recreación')),
    department VARCHAR(30) NOT NULL CHECK (department IN ('Tarija','Chuquisaca','La Paz','Cochabamba','Oruro','Potosí','Santa Cruz','Beni','Pando')),
    city VARCHAR(100) NOT NULL,
    address VARCHAR(250) NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    opening_hours VARCHAR(1000) NOT NULL DEFAULT '',
    admission_cents INTEGER NOT NULL DEFAULT 0 CHECK (admission_cents BETWEEN 0 AND 100000000),
    recommendations VARCHAR(3000) NOT NULL DEFAULT '',
    phone VARCHAR(30) NOT NULL DEFAULT '',
    manager_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    published BOOLEAN NOT NULL DEFAULT FALSE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((latitude IS NULL AND longitude IS NULL) OR (latitude IS NOT NULL AND longitude IS NOT NULL AND latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)),
    CHECK (NOT published OR status = 'active')
);
CREATE INDEX attractions_manager ON attractions(manager_id);
CREATE INDEX attractions_catalog ON attractions(status, published, department, category);
CREATE TRIGGER attractions_set_updated_at BEFORE UPDATE ON attractions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TABLE attraction_photos (
    id BIGSERIAL PRIMARY KEY,
    attraction_id BIGINT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
    position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 5),
    image BYTEA NOT NULL CHECK (octet_length(image) BETWEEN 1 AND 1048576)
);
CREATE INDEX attraction_photos_parent ON attraction_photos(attraction_id, position);
CREATE TABLE attraction_favorites (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attraction_id BIGINT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, attraction_id)
);
