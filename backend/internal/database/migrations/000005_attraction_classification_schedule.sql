-- Classification catalog copied from Andaria's categorias_seed.go.
CREATE TABLE attraction_categories (
 id BIGINT PRIMARY KEY,
 name VARCHAR(100) NOT NULL UNIQUE,
 position INTEGER NOT NULL
);
CREATE TABLE attraction_subcategories (
 id BIGINT PRIMARY KEY,
 category_id BIGINT NOT NULL REFERENCES attraction_categories(id),
 name VARCHAR(100) NOT NULL,
 position INTEGER NOT NULL,
 UNIQUE(category_id,name)
);
CREATE TABLE attraction_classifications (
 attraction_id BIGINT NOT NULL REFERENCES attractions(id) ON DELETE CASCADE,
 subcategory_id BIGINT NOT NULL REFERENCES attraction_subcategories(id),
 position INTEGER NOT NULL CHECK (position BETWEEN 0 AND 3),
 PRIMARY KEY(attraction_id,subcategory_id),
 UNIQUE(attraction_id,position)
);
CREATE INDEX attraction_classifications_filter ON attraction_classifications(subcategory_id, attraction_id);
-- Retain previous category text until the owner selects an exact subcategory.
ALTER TABLE attractions DROP CONSTRAINT attractions_category_check;
ALTER TABLE attractions
 ADD COLUMN schedule_mode VARCHAR(20) NOT NULL DEFAULT 'unspecified' CHECK (schedule_mode IN ('unspecified','scheduled','all_day')),
 ADD COLUMN opening_time VARCHAR(5) NOT NULL DEFAULT '',
 ADD COLUMN closing_time VARCHAR(5) NOT NULL DEFAULT '',
 ADD COLUMN opening_days JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(opening_days) = 'array' AND jsonb_array_length(opening_days) <= 7),
 ADD COLUMN season_mode VARCHAR(20) NOT NULL DEFAULT 'unspecified' CHECK (season_mode IN ('unspecified','all_year','months')),
 ADD COLUMN season_start_month INTEGER,
 ADD COLUMN season_end_month INTEGER,
 ADD CONSTRAINT attractions_season_check CHECK (
   (season_mode = 'months' AND season_start_month IS NOT NULL AND season_end_month IS NOT NULL AND season_start_month BETWEEN 1 AND 12 AND season_end_month BETWEEN 1 AND 12)
   OR (season_mode <> 'months' AND season_start_month IS NULL AND season_end_month IS NULL)
 ),
 ADD CONSTRAINT attractions_schedule_check CHECK (
   (schedule_mode = 'unspecified' AND opening_time = '' AND closing_time = '' AND jsonb_array_length(opening_days) = 0)
   OR (schedule_mode = 'all_day' AND opening_time = '' AND closing_time = '' AND jsonb_array_length(opening_days) > 0)
   OR (schedule_mode = 'scheduled' AND opening_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$' AND closing_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$' AND opening_time <> closing_time AND jsonb_array_length(opening_days) > 0)
 );
INSERT INTO attraction_categories(id,name,position) VALUES
(1,'Enoturismo',1),
(2,'Cultural',2),
(3,'Natural',3),
(4,'Deportivo',4);
INSERT INTO attraction_subcategories(id,category_id,name,position) VALUES
(1,1,'Bodegas',1),
(2,1,'Viñedos',2),
(3,1,'Catas',3),
(4,1,'Vendimia',4),
(5,1,'Vinoteca',5),
(6,2,'Arqueología',6),
(7,2,'Monumentos',7),
(8,2,'Mercados',8),
(9,2,'Museos',9),
(10,2,'Pueblos',10),
(11,2,'Festivales',11),
(12,2,'Espectáculos',12),
(13,2,'Histórico',13),
(14,3,'Montañas',14),
(15,3,'Lagos',15),
(16,3,'Ríos',16),
(17,3,'Cascadas',17),
(18,3,'Salares',18),
(19,3,'Desiertos',19),
(20,3,'Playas',20),
(21,3,'Fauna',21),
(22,3,'Flora',22),
(23,3,'Miradores',23),
(24,3,'Reservas protegidas',24),
(25,3,'Dunas',25),
(26,3,'Formaciones geológicas',26),
(27,3,'Posas',27),
(28,4,'Senderismo',28),
(29,4,'Escalada',29),
(30,4,'Ciclismo',30),
(31,4,'Acuáticos',31),
(32,4,'Aéreos',32),
(33,4,'Nieve',33);
