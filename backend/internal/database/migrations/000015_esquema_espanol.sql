-- Traduce identificadores de persistencia, sin cambiar datos ni contratos JSON.
-- Las migraciones 1 a 14 se conservan como historial inmutable.
DO $traduccion$
DECLARE
    tabla RECORD;
    columna RECORD;
    objeto RECORD;
    nombre_anterior TEXT;
    numero INTEGER;
BEGIN
    FOR tabla IN SELECT * FROM (VALUES
    ('users', 'usuarios'),
    ('auth_sessions', 'sesiones_autenticacion'),
    ('oauth_attempts', 'intentos_oauth'),
    ('agencies', 'agencias'),
    ('attractions', 'atracciones'),
    ('attraction_photos', 'fotos_atracciones'),
    ('attraction_favorites', 'favoritos_atracciones'),
    ('attraction_categories', 'categorias_atracciones'),
    ('attraction_subcategories', 'subcategorias_atracciones'),
    ('attraction_classifications', 'clasificaciones_atracciones'),
    ('tour_packages', 'paquetes_turisticos'),
    ('tour_package_photos', 'fotos_paquete'),
    ('tour_package_itinerary_days', 'dias_itinerario_paquete'),
    ('tour_package_itinerary_attractions', 'atracciones_itinerario_paquete'),
    ('tour_package_schedules', 'programaciones_paquete'),
    ('tour_package_departures', 'salidas_paquete'),
    ('tour_package_purchases', 'compras_paquete'),
    ('purchase_requests', 'solicitudes_compra'),
    ('pending_registrations', 'registros_pendientes'),
    ('schema_migrations', 'migraciones_esquema')
    ) AS nombres(anterior, nuevo)
    LOOP
        -- El registro de migraciones se traduce antes de ejecutar este archivo.
        IF tabla.anterior = 'schema_migrations' THEN CONTINUE; END IF;
        EXECUTE format('ALTER TABLE %I RENAME TO %I', tabla.anterior, tabla.nuevo);
        FOR columna IN SELECT * FROM (VALUES
    ('accepts_qr', 'acepta_qr'),
    ('accepts_transfer', 'acepta_transferencia'),
    ('account_holder', 'titular_cuenta'),
    ('account_number', 'numero_cuenta'),
    ('activities', 'actividades'),
    ('address', 'direccion'),
    ('admission_cents', 'entrada_centavos'),
    ('agency_id', 'agencia_id'),
    ('agency_review_user_id', 'usuario_revision_agencia_id'),
    ('attraction_id', 'atraccion_id'),
    ('bank_name', 'nombre_banco'),
    ('booking_closes_at', 'compras_cierran_en'),
    ('booking_cutoff_hours', 'horas_cierre_compras'),
    ('booking_opens_at', 'compras_abren_en'),
    ('bring', 'llevar'),
    ('cancellation_allowed', 'permite_cancelacion'),
    ('cancellation_notice_hours', 'horas_aviso_cancelacion'),
    ('cancellation_reason', 'motivo_cancelacion'),
    ('cancelled_at', 'cancelado_en'),
    ('capacity_count', 'cantidad_cupos'),
    ('category', 'categoria'),
    ('category_id', 'categoria_id'),
    ('city', 'ciudad'),
    ('closing_time', 'hora_cierre'),
    ('confirmed_capacity', 'cupos_confirmados'),
    ('created_at', 'creado_en'),
    ('day_number', 'numero_dia'),
    ('default_instructions', 'instrucciones_predeterminadas'),
    ('default_max_capacity', 'cupo_maximo_predeterminado'),
    ('default_meeting_point', 'punto_encuentro_predeterminado'),
    ('default_min_capacity', 'cupo_minimo_predeterminado'),
    ('department', 'departamento'),
    ('departure_id', 'salida_id'),
    ('departure_time', 'hora_salida'),
    ('description', 'descripcion'),
    ('difficulty', 'dificultad'),
    ('document_number', 'numero_documento'),
    ('duration_days', 'duracion_dias'),
    ('duration_nights', 'duracion_noches'),
    ('email', 'correo'),
    ('email_verified_at', 'correo_verificado_en'),
    ('excludes', 'no_incluye'),
    ('expires_at', 'vence_en'),
    ('failed_login_attempts', 'intentos_fallidos_acceso'),
    ('first_name', 'nombres'),
    ('foreign_adults', 'adultos_extranjeros'),
    ('foreign_surcharge_cents', 'recargo_extranjero_centavos'),
    ('free_minor_count', 'cantidad_menores_gratis'),
    ('frequency_type', 'tipo_frecuencia'),
    ('google_subject', 'identificador_google'),
    ('held_capacity', 'cupos_reservados'),
    ('image', 'imagen'),
    ('includes', 'incluye'),
    ('instructions', 'instrucciones'),
    ('ip_address', 'direccion_ip'),
    ('is_exception', 'es_excepcion'),
    ('itinerary_day_id', 'dia_itinerario_id'),
    ('last_activity_at', 'ultima_actividad_en'),
    ('last_login_at', 'ultimo_acceso_en'),
    ('last_name', 'apellidos'),
    ('latitude', 'latitud'),
    ('locked_until', 'bloqueado_hasta'),
    ('longitude', 'longitud'),
    ('manager_id', 'encargado_id'),
    ('max_capacity', 'cupo_maximo'),
    ('maximum_advance_days', 'dias_maxima_anticipacion'),
    ('meeting_at', 'encuentro_en'),
    ('meeting_point', 'punto_encuentro'),
    ('meeting_time', 'hora_encuentro'),
    ('min_capacity', 'cupo_minimo'),
    ('minimum_paying_age', 'edad_minima_pago'),
    ('minimum_review_at', 'revision_minimo_en'),
    ('minimum_review_user_id', 'usuario_revision_minimo_id'),
    ('minimum_reviewed_at', 'minimo_revisado_en'),
    ('minors', 'menores'),
    ('modified_by_user_id', 'usuario_modificacion_id'),
    ('name', 'nombre'),
    ('national_adults', 'adultos_nacionales'),
    ('national_price_cents', 'precio_nacional_centavos'),
    ('national_unit_price_cents', 'precio_unitario_nacional_centavos'),
    ('nationality', 'nacionalidad'),
    ('opening_days', 'dias_apertura'),
    ('opening_hours', 'horario_atencion'),
    ('opening_time', 'hora_apertura'),
    ('package_id', 'paquete_id'),
    ('password_hash', 'hash_contrasena'),
    ('paying_minor_count', 'cantidad_menores_pago'),
    ('payload_hash', 'hash_contenido'),
    ('payment_instructions', 'instrucciones_pago'),
    ('payment_method', 'metodo_pago'),
    ('payment_proof', 'comprobante_pago'),
    ('phone', 'telefono'),
    ('position', 'posicion'),
    ('published', 'publicado'),
    ('purchase_id', 'compra_id'),
    ('qr_image', 'imagen_qr'),
    ('recommendations', 'recomendaciones'),
    ('reference', 'referencia'),
    ('refund_account_holder', 'titular_cuenta_reembolso'),
    ('refund_account_number', 'numero_cuenta_reembolso'),
    ('refund_bank_name', 'nombre_banco_reembolso'),
    ('refund_completed_by_user_id', 'usuario_reembolso_id'),
    ('refund_due_at', 'reembolso_vence_en'),
    ('refund_method', 'metodo_reembolso'),
    ('refund_proof', 'comprobante_reembolso'),
    ('refund_qr', 'qr_reembolso'),
    ('refund_reason', 'motivo_reembolso'),
    ('refund_reference', 'referencia_reembolso'),
    ('refund_requested_at', 'reembolso_solicitado_en'),
    ('refunded_at', 'reembolsado_en'),
    ('rejection_reason', 'motivo_rechazo'),
    ('request_key', 'clave_solicitud'),
    ('reviewed_at', 'revisado_en'),
    ('revoked_at', 'revocado_en'),
    ('role', 'rol'),
    ('schedule_id', 'programacion_id'),
    ('schedule_mode', 'modo_horario'),
    ('season_end_month', 'mes_fin_temporada'),
    ('season_mode', 'modo_temporada'),
    ('season_start_month', 'mes_inicio_temporada'),
    ('session_id', 'sesion_id'),
    ('starts_at', 'salida_en'),
    ('state_hash', 'hash_estado'),
    ('status', 'estado'),
    ('subcategory_id', 'subcategoria_id'),
    ('title', 'titulo'),
    ('token_hash', 'hash_token'),
    ('total_cents', 'total_centavos'),
    ('tourist_document', 'documento_turista'),
    ('tourist_email', 'correo_turista'),
    ('tourist_id', 'turista_id'),
    ('tourist_name', 'nombre_turista'),
    ('tourist_nationality', 'nacionalidad_turista'),
    ('tourist_phone', 'telefono_turista'),
    ('updated_at', 'actualizado_en'),
    ('user_agent', 'agente_usuario'),
    ('user_id', 'usuario_id'),
    ('valid_from', 'vigente_desde'),
    ('valid_until', 'vigente_hasta'),
    ('verifier', 'verificador'),
    ('weekdays', 'dias_semana'),
    ('applied_at', 'aplicado_en')
        ) AS nombres(anterior, nuevo)
        LOOP
            IF EXISTS (SELECT 1 FROM information_schema.columns
                WHERE table_schema = current_schema() AND table_name = tabla.nuevo
                AND column_name = columna.anterior) THEN
                EXECUTE format('ALTER TABLE %I RENAME COLUMN %I TO %I',
                    tabla.nuevo, columna.anterior, columna.nuevo);
            END IF;
        END LOOP;
        -- PostgreSQL mantiene las referencias y expresiones al renombrar columnas.
        -- Se normalizan tambien los nombres de restricciones, indices y secuencias.
        numero := 0;
        FOR objeto IN SELECT conname FROM pg_constraint
            WHERE conrelid = to_regclass(tabla.nuevo) ORDER BY conname
        LOOP
            numero := numero + 1;
            EXECUTE format('ALTER TABLE %I RENAME CONSTRAINT %I TO %I',
                tabla.nuevo, objeto.conname, tabla.nuevo || '_restriccion_' || numero);
        END LOOP;
        numero := 0;
        FOR objeto IN SELECT c.relname FROM pg_index i JOIN pg_class c ON c.oid = i.indexrelid
            WHERE i.indrelid = to_regclass(tabla.nuevo)
            AND NOT EXISTS (SELECT 1 FROM pg_constraint r WHERE r.conindid = i.indexrelid)
            ORDER BY c.relname
        LOOP
            numero := numero + 1;
            EXECUTE format('ALTER INDEX %I RENAME TO %I', objeto.relname,
                tabla.nuevo || '_indice_' || numero);
        END LOOP;
        nombre_anterior := NULL;
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema()
            AND table_name = tabla.nuevo AND column_name = 'id') THEN
            nombre_anterior := pg_get_serial_sequence(tabla.nuevo, 'id');
        END IF;
        IF nombre_anterior IS NOT NULL THEN
            EXECUTE format('ALTER SEQUENCE %s RENAME TO %I', nombre_anterior,
                tabla.nuevo || '_id_secuencia');
        END IF;
        FOR objeto IN SELECT tgname FROM pg_trigger
            WHERE tgrelid = to_regclass(tabla.nuevo) AND NOT tgisinternal
        LOOP
            EXECUTE format('ALTER TRIGGER %I ON %I RENAME TO %I', objeto.tgname,
                tabla.nuevo, tabla.nuevo || '_actualizar_fecha');
        END LOOP;
    END LOOP;
END;
$traduccion$;

-- Los cuerpos PL/pgSQL no se reescriben al renombrar columnas.
ALTER FUNCTION set_updated_at() RENAME TO establecer_actualizado_en;
CREATE OR REPLACE FUNCTION establecer_actualizado_en()
RETURNS TRIGGER AS $$
BEGIN
    NEW.actualizado_en = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
