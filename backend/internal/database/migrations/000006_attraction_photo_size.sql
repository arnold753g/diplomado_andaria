ALTER TABLE attraction_photos DROP CONSTRAINT attraction_photos_image_check;
ALTER TABLE attraction_photos ADD CONSTRAINT attraction_photos_image_check CHECK (octet_length(image) BETWEEN 1 AND 5242880);
