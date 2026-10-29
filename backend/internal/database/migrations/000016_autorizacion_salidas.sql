ALTER TABLE salidas_paquete
    ADD COLUMN autorizado_en TIMESTAMPTZ,
    ADD COLUMN usuario_autorizacion_id BIGINT REFERENCES usuarios(id) ON DELETE RESTRICT,
    ADD COLUMN autorizacion_revocada_en TIMESTAMPTZ,
    ADD COLUMN motivo_revocacion_autorizacion VARCHAR(250) NOT NULL DEFAULT '',
    ADD CONSTRAINT salidas_paquete_autorizacion_responsable CHECK (
        (autorizado_en IS NULL AND usuario_autorizacion_id IS NULL AND autorizacion_revocada_en IS NULL)
        OR (autorizado_en IS NOT NULL AND usuario_autorizacion_id IS NOT NULL)
    ),
    ADD CONSTRAINT salidas_paquete_autorizacion_fechas CHECK (
        autorizacion_revocada_en IS NULL OR autorizacion_revocada_en >= autorizado_en
    );

-- La autorizacion operativa es independiente del estado comercial y del minimo.
-- Se conserva la ultima decision y su responsable aunque pierda vigencia.
CREATE FUNCTION validar_autorizacion_salida() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.autorizado_en IS NOT NULL AND OLD.autorizacion_revocada_en IS NULL
       AND NEW.autorizacion_revocada_en IS NULL THEN
        IF NEW.estado = 'cancelled' THEN
            NEW.autorizacion_revocada_en = clock_timestamp();
            NEW.motivo_revocacion_autorizacion = 'La salida fue cancelada';
        ELSIF NEW.salida_en > NOW() AND NEW.cupos_confirmados < NEW.cupo_minimo THEN
            NEW.autorizacion_revocada_en = clock_timestamp();
            NEW.motivo_revocacion_autorizacion = 'Los cupos confirmados quedaron por debajo del minimo';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER salidas_paquete_validar_autorizacion
    BEFORE UPDATE ON salidas_paquete
    FOR EACH ROW EXECUTE FUNCTION validar_autorizacion_salida();
CREATE INDEX salidas_paquete_autorizacion_pendiente
    ON salidas_paquete (salida_en) WHERE autorizado_en IS NULL OR autorizacion_revocada_en IS NOT NULL;
