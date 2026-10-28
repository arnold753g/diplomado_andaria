ALTER TABLE compras_paquete
    ADD COLUMN datos_comprobante JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN pago_confirmado_en TIMESTAMPTZ;

-- Los datos personales y las tarifas ya se conservan en la compra.
-- Esta copia preserva los datos comerciales que antes se consultaban del catalogo.
CREATE FUNCTION compras_paquete_datos_comprobante(paquete BIGINT, salida BIGINT, agencia BIGINT, origen TEXT)
RETURNS JSONB LANGUAGE SQL AS $$
    SELECT jsonb_build_object(
        'origen', origen, 'registrado_en', clock_timestamp(),
        'agencia_nombre', a.nombre, 'agencia_direccion', a.direccion,
        'agencia_telefono', a.telefono, 'agencia_correo', a.correo,
        'paquete_nombre', p.nombre, 'duracion_dias', p.duracion_dias,
        'duracion_noches', p.duracion_noches, 'llevar', p.llevar,
        'permite_cancelacion', p.permite_cancelacion,
        'horas_aviso_cancelacion', p.horas_aviso_cancelacion,
        'salida_en', s.salida_en, 'encuentro_en', s.encuentro_en,
        'punto_encuentro', s.punto_encuentro
    ) FROM paquetes_turisticos p
      JOIN agencias a ON a.id = agencia
      JOIN salidas_paquete s ON s.id = salida
    WHERE p.id = paquete AND s.paquete_id = p.id AND p.agencia_id = a.id;
$$;

-- No se atribuyen datos historicos que la version anterior no guardaba.
UPDATE compras_paquete SET
    datos_comprobante = compras_paquete_datos_comprobante(paquete_id, salida_id, agencia_id, 'migracion'),
    pago_confirmado_en = CASE WHEN estado = 'confirmed' THEN revisado_en ELSE NULL END;

CREATE FUNCTION compras_paquete_conservar_comprobante() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        NEW.datos_comprobante = compras_paquete_datos_comprobante(NEW.paquete_id, NEW.salida_id, NEW.agencia_id, 'compra');
        NEW.pago_confirmado_en = CASE WHEN NEW.estado = 'confirmed' THEN NEW.revisado_en ELSE NULL END;
    ELSE
        NEW.datos_comprobante = OLD.datos_comprobante;
        NEW.pago_confirmado_en = OLD.pago_confirmado_en;
        IF OLD.pago_confirmado_en IS NULL AND NEW.estado = 'confirmed' AND OLD.estado <> 'confirmed' THEN
            NEW.pago_confirmado_en = COALESCE(NEW.revisado_en, clock_timestamp());
        END IF;
    END IF;
    RETURN NEW;
END;
$$;
CREATE TRIGGER compras_paquete_conservar_comprobante
    BEFORE INSERT OR UPDATE ON compras_paquete
    FOR EACH ROW EXECUTE FUNCTION compras_paquete_conservar_comprobante();
