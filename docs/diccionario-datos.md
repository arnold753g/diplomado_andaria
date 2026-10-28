# Diccionario físico — Andaria

**Ampliación del 28/10/2026:** la [migración 17 de comprobantes PDF](comprobantes-pdf.md) agrega `compras_paquete.datos_comprobante` (JSONB, no nulo, copia comercial privada) y `compras_paquete.pago_confirmado_en` (TIMESTAMPTZ, nullable), con funciones y disparador para conservar la copia y la primera confirmación. El inventario histórico siguiente no incorpora esas dos columnas; no se agregan tablas.

**Ampliacion del 27/10/2026:** el corte del catalogo mostrado abajo corresponde a la migracion 15. La [migracion 16 de Gestionar pagos](gestion-pagos.md) agrega cuatro columnas de autorizacion a `salidas_paquete`, sus restricciones, indice y disparador, manteniendo 20 tablas.

Esquema observado el 26/10/2026 en PostgreSQL 17.5 local aislado, aplicando las migraciones 1–15. Contiene 19 tablas de aplicación y `migraciones_esquema`: 20 tablas, 229 columnas y 140 restricciones. Producción declara PostgreSQL 16; no se consultó ni modificó el VPS.

Identificadores en español, sin tildes ni ñ, en `snake_case`. Los estados, claves de documentos JSON y contratos HTTP conservan sus valores actuales. El registro SQL histórico 1–14 permanece en inglés; la migración 15 traduce la estructura resultante.

[Evidencia del esquema](evidencias/esquema-espanol-2026-10-26/esquema-fisico.json) · [Despliegue y compatibilidad](base-datos-espanol.md)

## agencias

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('agencias_id_secuencia'::regclass) | No |
| nombre | character varying(160) | No | Sin valor | No |
| descripcion | text | No | ''::text | No |
| departamento | character varying(30) | No | 'Tarija'::character varying | No |
| ciudad | character varying(100) | No | Sin valor | No |
| direccion | character varying(250) | No | Sin valor | No |
| telefono | character varying(30) | No | Sin valor | Sí |
| correo | character varying(320) | No | Sin valor | Sí |
| encargado_id | bigint | No | Sin valor | No |
| estado | character varying(20) | No | 'active'::character varying | No |
| publicado | boolean | No | false | No |
| edad_minima_pago | integer | No | 6 | No |
| acepta_qr | boolean | No | false | No |
| acepta_transferencia | boolean | No | false | No |
| nombre_banco | character varying(100) | No | ''::character varying | No |
| titular_cuenta | character varying(160) | No | ''::character varying | Sí |
| numero_cuenta | character varying(50) | No | ''::character varying | Sí |
| instrucciones_pago | character varying(1000) | No | ''::character varying | No |
| imagen_qr | bytea | Sí | Sin valor | No |
| version | integer | No | 1 | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `agencias_restriccion_1`: `CHECK (((NOT acepta_qr) OR ((octet_length(imagen_qr) > 0) AND (imagen_qr IS NOT NULL))))`
- `agencias_restriccion_2`: `CHECK (((NOT acepta_transferencia) OR ((length(TRIM(BOTH FROM nombre_banco)) > 0) AND (length(TRIM(BOTH FROM titular_cuenta)) > 0) AND (length(TRIM(BOTH FROM numero_cuenta)) > 0))))`
- `agencias_restriccion_3`: `CHECK (((departamento)::text = ANY ((ARRAY['Tarija'::character varying, 'Chuquisaca'::character varying, 'La Paz'::character varying, 'Cochabamba'::character varying, 'Oruro'::character varying, 'PotosÃ­'::character varying, 'Santa Cruz'::character varying, 'Beni'::character varying, 'Pando'::character varying])::text[])))`
- `agencias_restriccion_4`: `FOREIGN KEY (encargado_id) REFERENCES usuarios(id) ON DELETE RESTRICT`
- `agencias_restriccion_5`: `UNIQUE (encargado_id)`
- `agencias_restriccion_6`: `CHECK (((edad_minima_pago >= 0) AND (edad_minima_pago <= 18)))`
- `agencias_restriccion_7`: `PRIMARY KEY (id)`
- `agencias_restriccion_8`: `CHECK (((estado)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying])::text[])))`
- `agencias_restriccion_9`: `CHECK ((version > 0))`

## atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('atracciones_id_secuencia'::regclass) | No |
| nombre | character varying(160) | No | Sin valor | No |
| descripcion | text | No | ''::text | No |
| categoria | character varying(30) | No | Sin valor | No |
| departamento | character varying(30) | No | Sin valor | No |
| ciudad | character varying(100) | No | Sin valor | No |
| direccion | character varying(250) | No | Sin valor | No |
| latitud | double precision | Sí | Sin valor | No |
| longitud | double precision | Sí | Sin valor | No |
| horario_atencion | character varying(1000) | No | ''::character varying | No |
| entrada_centavos | integer | No | 0 | No |
| recomendaciones | character varying(3000) | No | ''::character varying | No |
| telefono | character varying(30) | No | ''::character varying | Sí |
| encargado_id | bigint | No | Sin valor | No |
| estado | character varying(20) | No | 'active'::character varying | No |
| publicado | boolean | No | false | No |
| version | integer | No | 1 | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |
| modo_horario | character varying(20) | No | 'unspecified'::character varying | No |
| hora_apertura | character varying(5) | No | ''::character varying | No |
| hora_cierre | character varying(5) | No | ''::character varying | No |
| dias_apertura | jsonb | No | '[]'::jsonb | No |
| modo_temporada | character varying(20) | No | 'unspecified'::character varying | No |
| mes_inicio_temporada | integer | Sí | Sin valor | No |
| mes_fin_temporada | integer | Sí | Sin valor | No |

Restricciones:

- `atracciones_restriccion_1`: `CHECK (((entrada_centavos >= 0) AND (entrada_centavos <= 100000000)))`
- `atracciones_restriccion_10`: `CHECK (((((modo_temporada)::text = 'months'::text) AND (mes_inicio_temporada IS NOT NULL) AND (mes_fin_temporada IS NOT NULL) AND ((mes_inicio_temporada >= 1) AND (mes_inicio_temporada <= 12)) AND ((mes_fin_temporada >= 1) AND (mes_fin_temporada <= 12))) OR (((modo_temporada)::text <> 'months'::text) AND (mes_inicio_temporada IS NULL) AND (mes_fin_temporada IS NULL))))`
- `atracciones_restriccion_11`: `CHECK (((modo_temporada)::text = ANY ((ARRAY['unspecified'::character varying, 'all_year'::character varying, 'months'::character varying])::text[])))`
- `atracciones_restriccion_12`: `CHECK (((estado)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying])::text[])))`
- `atracciones_restriccion_13`: `CHECK ((version > 0))`
- `atracciones_restriccion_2`: `CHECK ((((latitud IS NULL) AND (longitud IS NULL)) OR ((latitud IS NOT NULL) AND (longitud IS NOT NULL) AND ((latitud >= ('-90'::integer)::double precision) AND (latitud <= (90)::double precision)) AND ((longitud >= ('-180'::integer)::double precision) AND (longitud <= (180)::double precision)))))`
- `atracciones_restriccion_3`: `CHECK (((NOT publicado) OR ((estado)::text = 'active'::text)))`
- `atracciones_restriccion_4`: `CHECK (((departamento)::text = ANY ((ARRAY['Tarija'::character varying, 'Chuquisaca'::character varying, 'La Paz'::character varying, 'Cochabamba'::character varying, 'Oruro'::character varying, 'PotosÃ­'::character varying, 'Santa Cruz'::character varying, 'Beni'::character varying, 'Pando'::character varying])::text[])))`
- `atracciones_restriccion_5`: `FOREIGN KEY (encargado_id) REFERENCES usuarios(id) ON DELETE RESTRICT`
- `atracciones_restriccion_6`: `CHECK (((jsonb_typeof(dias_apertura) = 'array'::text) AND (jsonb_array_length(dias_apertura) <= 7)))`
- `atracciones_restriccion_7`: `PRIMARY KEY (id)`
- `atracciones_restriccion_8`: `CHECK (((((modo_horario)::text = 'unspecified'::text) AND ((hora_apertura)::text = ''::text) AND ((hora_cierre)::text = ''::text) AND (jsonb_array_length(dias_apertura) = 0)) OR (((modo_horario)::text = 'all_day'::text) AND ((hora_apertura)::text = ''::text) AND ((hora_cierre)::text = ''::text) AND (jsonb_array_length(dias_apertura) > 0)) OR (((modo_horario)::text = 'scheduled'::text) AND ((hora_apertura)::text ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'::text) AND ((hora_cierre)::text ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'::text) AND ((hora_apertura)::text <> (hora_cierre)::text) AND (jsonb_array_length(dias_apertura) > 0))))`
- `atracciones_restriccion_9`: `CHECK (((modo_horario)::text = ANY ((ARRAY['unspecified'::character varying, 'scheduled'::character varying, 'all_day'::character varying])::text[])))`

## atracciones_itinerario_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| dia_itinerario_id | bigint | No | Sin valor | No |
| atraccion_id | bigint | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |

Restricciones:

- `atracciones_itinerario_paquete_restriccion_1`: `UNIQUE (dia_itinerario_id, posicion)`
- `atracciones_itinerario_paquete_restriccion_2`: `FOREIGN KEY (atraccion_id) REFERENCES atracciones(id) ON DELETE RESTRICT`
- `atracciones_itinerario_paquete_restriccion_3`: `FOREIGN KEY (dia_itinerario_id) REFERENCES dias_itinerario_paquete(id) ON DELETE CASCADE`
- `atracciones_itinerario_paquete_restriccion_4`: `PRIMARY KEY (dia_itinerario_id, atraccion_id)`
- `atracciones_itinerario_paquete_restriccion_5`: `CHECK (((posicion >= 0) AND (posicion <= 19)))`

## categorias_atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | Sin valor | No |
| nombre | character varying(100) | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |

Restricciones:

- `categorias_atracciones_restriccion_1`: `UNIQUE (nombre)`
- `categorias_atracciones_restriccion_2`: `PRIMARY KEY (id)`

## clasificaciones_atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| atraccion_id | bigint | No | Sin valor | No |
| subcategoria_id | bigint | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |

Restricciones:

- `clasificaciones_atracciones_restriccion_1`: `FOREIGN KEY (atraccion_id) REFERENCES atracciones(id) ON DELETE CASCADE`
- `clasificaciones_atracciones_restriccion_2`: `UNIQUE (atraccion_id, posicion)`
- `clasificaciones_atracciones_restriccion_3`: `PRIMARY KEY (atraccion_id, subcategoria_id)`
- `clasificaciones_atracciones_restriccion_4`: `CHECK (((posicion >= 0) AND (posicion <= 3)))`
- `clasificaciones_atracciones_restriccion_5`: `FOREIGN KEY (subcategoria_id) REFERENCES subcategorias_atracciones(id)`

## compras_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('compras_paquete_id_secuencia'::regclass) | No |
| referencia | character varying(40) | No | Sin valor | No |
| turista_id | bigint | No | Sin valor | No |
| agencia_id | bigint | No | Sin valor | No |
| paquete_id | bigint | No | Sin valor | No |
| salida_id | bigint | No | Sin valor | No |
| estado | character varying(24) | No | 'payment_review'::character varying | No |
| metodo_pago | character varying(16) | No | Sin valor | No |
| adultos_nacionales | integer | No | 0 | No |
| adultos_extranjeros | integer | No | 0 | No |
| menores | jsonb | No | '[]'::jsonb | No |
| cantidad_menores_gratis | integer | No | 0 | No |
| cantidad_menores_pago | integer | No | 0 | No |
| cantidad_cupos | integer | No | Sin valor | No |
| precio_unitario_nacional_centavos | integer | No | Sin valor | No |
| recargo_extranjero_centavos | integer | No | Sin valor | No |
| total_centavos | integer | No | Sin valor | No |
| nombre_turista | character varying(210) | No | Sin valor | No |
| correo_turista | character varying(320) | No | Sin valor | Sí |
| telefono_turista | character varying(30) | No | ''::character varying | Sí |
| documento_turista | character varying(40) | No | ''::character varying | Sí |
| nacionalidad_turista | character varying(80) | No | ''::character varying | No |
| comprobante_pago | bytea | No | Sin valor | Sí |
| usuario_revision_agencia_id | bigint | Sí | Sin valor | No |
| revisado_en | timestamp with time zone | Sí | Sin valor | No |
| motivo_rechazo | text | No | ''::text | No |
| version | integer | No | 1 | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |
| cancelado_en | timestamp with time zone | Sí | Sin valor | No |
| motivo_cancelacion | text | No | ''::text | No |
| motivo_reembolso | text | No | ''::text | No |
| reembolso_solicitado_en | timestamp with time zone | Sí | Sin valor | No |
| reembolso_vence_en | timestamp with time zone | Sí | Sin valor | No |
| metodo_reembolso | character varying(20) | Sí | Sin valor | No |
| qr_reembolso | bytea | Sí | Sin valor | Sí |
| nombre_banco_reembolso | character varying(120) | No | ''::character varying | No |
| titular_cuenta_reembolso | character varying(160) | No | ''::character varying | Sí |
| numero_cuenta_reembolso | character varying(80) | No | ''::character varying | Sí |
| comprobante_reembolso | bytea | Sí | Sin valor | Sí |
| referencia_reembolso | character varying(160) | No | ''::character varying | No |
| reembolsado_en | timestamp with time zone | Sí | Sin valor | No |
| usuario_reembolso_id | bigint | Sí | Sin valor | No |

Restricciones:

- `compras_paquete_restriccion_1`: `FOREIGN KEY (agencia_id) REFERENCES agencias(id) ON DELETE RESTRICT`
- `compras_paquete_restriccion_10`: `CHECK (((adultos_nacionales >= 0) AND (adultos_nacionales <= 500)))`
- `compras_paquete_restriccion_11`: `CHECK ((precio_unitario_nacional_centavos >= 0))`
- `compras_paquete_restriccion_12`: `FOREIGN KEY (paquete_id) REFERENCES paquetes_turisticos(id) ON DELETE RESTRICT`
- `compras_paquete_restriccion_13`: `CHECK (((cantidad_menores_pago >= 0) AND (cantidad_menores_pago <= 500)))`
- `compras_paquete_restriccion_14`: `CHECK (((metodo_pago)::text = ANY ((ARRAY['qr'::character varying, 'transfer'::character varying])::text[])))`
- `compras_paquete_restriccion_15`: `CHECK (((octet_length(comprobante_pago) >= 1) AND (octet_length(comprobante_pago) <= 5242880)))`
- `compras_paquete_restriccion_16`: `PRIMARY KEY (id)`
- `compras_paquete_restriccion_17`: `UNIQUE (referencia)`
- `compras_paquete_restriccion_18`: `FOREIGN KEY (usuario_reembolso_id) REFERENCES usuarios(id) ON DELETE SET NULL`
- `compras_paquete_restriccion_19`: `CHECK (((metodo_reembolso IS NULL) OR ((metodo_reembolso)::text = ''::text) OR (((metodo_reembolso)::text = 'qr'::text) AND (qr_reembolso IS NOT NULL) AND ((octet_length(qr_reembolso) >= 1) AND (octet_length(qr_reembolso) <= 5242880))) OR (((metodo_reembolso)::text = 'bank_transfer'::text) AND (length(TRIM(BOTH FROM nombre_banco_reembolso)) >= 2) AND (length(TRIM(BOTH FROM titular_cuenta_reembolso)) >= 3) AND (length(TRIM(BOTH FROM numero_cuenta_reembolso)) >= 3))))`
- `compras_paquete_restriccion_2`: `FOREIGN KEY (usuario_revision_agencia_id) REFERENCES usuarios(id) ON DELETE SET NULL`
- `compras_paquete_restriccion_20`: `CHECK (((reembolso_vence_en IS NULL) OR ((reembolso_solicitado_en IS NOT NULL) AND (reembolso_vence_en > reembolso_solicitado_en))))`
- `compras_paquete_restriccion_21`: `CHECK (((metodo_reembolso IS NULL) OR ((metodo_reembolso)::text = ANY ((ARRAY[''::character varying, 'qr'::character varying, 'bank_transfer'::character varying])::text[]))))`
- `compras_paquete_restriccion_22`: `CHECK (((comprobante_reembolso IS NULL) OR ((octet_length(comprobante_reembolso) >= 1) AND (octet_length(comprobante_reembolso) <= 5242880))))`
- `compras_paquete_restriccion_23`: `CHECK ((((estado)::text <> 'refunded'::text) OR ((metodo_reembolso)::text = ANY ((ARRAY['qr'::character varying, 'bank_transfer'::character varying])::text[]))))`
- `compras_paquete_restriccion_24`: `CHECK (((((estado)::text = 'payment_review'::text) AND (revisado_en IS NULL) AND (motivo_rechazo = ''::text)) OR (((estado)::text = 'correction_requested'::text) AND (revisado_en IS NOT NULL) AND (length(TRIM(BOTH FROM motivo_rechazo)) >= 5)) OR (((estado)::text = 'confirmed'::text) AND (revisado_en IS NOT NULL) AND (motivo_rechazo = ''::text)) OR (((estado)::text = 'payment_rejected'::text) AND (revisado_en IS NOT NULL) AND (length(TRIM(BOTH FROM motivo_rechazo)) >= 5)) OR (((estado)::text = 'refund_pending'::text) AND (reembolso_solicitado_en IS NOT NULL) AND (reembolso_vence_en IS NOT NULL) AND (length(TRIM(BOTH FROM motivo_reembolso)) >= 5)) OR (((estado)::text = 'refunded'::text) AND (reembolso_solicitado_en IS NOT NULL) AND (reembolso_vence_en IS NOT NULL) AND (reembolsado_en IS NOT NULL) AND (comprobante_reembolso IS NOT NULL)) OR ((estado)::text = 'cancelled'::text)))`
- `compras_paquete_restriccion_25`: `CHECK (((estado)::text = ANY ((ARRAY['payment_review'::character varying, 'correction_requested'::character varying, 'confirmed'::character varying, 'payment_rejected'::character varying, 'cancelled'::character varying, 'refund_pending'::character varying, 'refunded'::character varying])::text[])))`
- `compras_paquete_restriccion_26`: `CHECK ((total_centavos > 0))`
- `compras_paquete_restriccion_27`: `FOREIGN KEY (turista_id) REFERENCES usuarios(id) ON DELETE RESTRICT`
- `compras_paquete_restriccion_28`: `CHECK ((version > 0))`
- `compras_paquete_restriccion_3`: `CHECK (((cantidad_cupos >= 1) AND (cantidad_cupos <= 500)))`
- `compras_paquete_restriccion_4`: `CHECK (((adultos_nacionales + adultos_extranjeros) >= 1))`
- `compras_paquete_restriccion_5`: `CHECK ((cantidad_cupos >= (adultos_nacionales + adultos_extranjeros)))`
- `compras_paquete_restriccion_6`: `FOREIGN KEY (salida_id) REFERENCES salidas_paquete(id) ON DELETE RESTRICT`
- `compras_paquete_restriccion_7`: `CHECK (((adultos_extranjeros >= 0) AND (adultos_extranjeros <= 500)))`
- `compras_paquete_restriccion_8`: `CHECK ((recargo_extranjero_centavos >= 0))`
- `compras_paquete_restriccion_9`: `CHECK (((cantidad_menores_gratis >= 0) AND (cantidad_menores_gratis <= 500)))`

## dias_itinerario_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('dias_itinerario_paquete_id_secuencia'::regclass) | No |
| paquete_id | bigint | No | Sin valor | No |
| numero_dia | integer | No | Sin valor | No |
| titulo | character varying(160) | No | Sin valor | No |
| descripcion | text | No | ''::text | No |
| actividades | jsonb | No | '[]'::jsonb | No |

Restricciones:

- `dias_itinerario_paquete_restriccion_1`: `CHECK (((numero_dia >= 1) AND (numero_dia <= 30)))`
- `dias_itinerario_paquete_restriccion_2`: `UNIQUE (paquete_id, numero_dia)`
- `dias_itinerario_paquete_restriccion_3`: `FOREIGN KEY (paquete_id) REFERENCES paquetes_turisticos(id) ON DELETE CASCADE`
- `dias_itinerario_paquete_restriccion_4`: `PRIMARY KEY (id)`

## favoritos_atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| usuario_id | bigint | No | Sin valor | No |
| atraccion_id | bigint | No | Sin valor | No |
| creado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `favoritos_atracciones_restriccion_1`: `FOREIGN KEY (atraccion_id) REFERENCES atracciones(id) ON DELETE CASCADE`
- `favoritos_atracciones_restriccion_2`: `PRIMARY KEY (usuario_id, atraccion_id)`
- `favoritos_atracciones_restriccion_3`: `FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE`

## fotos_atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('fotos_atracciones_id_secuencia'::regclass) | No |
| atraccion_id | bigint | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |
| imagen | bytea | No | Sin valor | No |

Restricciones:

- `fotos_atracciones_restriccion_1`: `FOREIGN KEY (atraccion_id) REFERENCES atracciones(id) ON DELETE CASCADE`
- `fotos_atracciones_restriccion_2`: `CHECK (((octet_length(imagen) >= 1) AND (octet_length(imagen) <= 5242880)))`
- `fotos_atracciones_restriccion_3`: `PRIMARY KEY (id)`
- `fotos_atracciones_restriccion_4`: `CHECK (((posicion >= 0) AND (posicion <= 5)))`

## fotos_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('fotos_paquete_id_secuencia'::regclass) | No |
| paquete_id | bigint | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |
| imagen | bytea | No | Sin valor | No |

Restricciones:

- `fotos_paquete_restriccion_1`: `CHECK (((octet_length(imagen) >= 1) AND (octet_length(imagen) <= 5242880)))`
- `fotos_paquete_restriccion_2`: `FOREIGN KEY (paquete_id) REFERENCES paquetes_turisticos(id) ON DELETE CASCADE`
- `fotos_paquete_restriccion_3`: `PRIMARY KEY (id)`
- `fotos_paquete_restriccion_4`: `CHECK (((posicion >= 0) AND (posicion <= 5)))`

## intentos_oauth

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| hash_estado | character varying(64) | No | Sin valor | Sí |
| verificador | character varying(128) | No | Sin valor | Sí |
| usuario_id | bigint | Sí | Sin valor | No |
| sesion_id | character varying(64) | Sí | Sin valor | No |
| vence_en | timestamp with time zone | No | Sin valor | No |

Restricciones:

- `intentos_oauth_restriccion_1`: `PRIMARY KEY (hash_estado)`
- `intentos_oauth_restriccion_2`: `FOREIGN KEY (sesion_id) REFERENCES sesiones_autenticacion(id) ON DELETE CASCADE`
- `intentos_oauth_restriccion_3`: `FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE`

## migraciones_esquema

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| version | bigint | No | Sin valor | No |
| nombre | text | No | Sin valor | No |
| aplicado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `migraciones_esquema_clave_primaria`: `PRIMARY KEY (version)`

## paquetes_turisticos

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('paquetes_turisticos_id_secuencia'::regclass) | No |
| agencia_id | bigint | No | Sin valor | No |
| nombre | character varying(160) | No | Sin valor | No |
| descripcion | text | No | ''::text | No |
| duracion_dias | integer | No | 1 | No |
| duracion_noches | integer | No | 0 | No |
| dificultad | character varying(20) | No | ''::character varying | No |
| precio_nacional_centavos | integer | No | 0 | No |
| recargo_extranjero_centavos | integer | No | 0 | No |
| incluye | jsonb | No | '[]'::jsonb | No |
| no_incluye | jsonb | No | '[]'::jsonb | No |
| llevar | jsonb | No | '[]'::jsonb | No |
| permite_cancelacion | boolean | No | false | No |
| horas_aviso_cancelacion | integer | No | 0 | No |
| publicado | boolean | No | false | No |
| version | integer | No | 1 | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `paquetes_turisticos_restriccion_1`: `FOREIGN KEY (agencia_id) REFERENCES agencias(id) ON DELETE RESTRICT`
- `paquetes_turisticos_restriccion_10`: `PRIMARY KEY (id)`
- `paquetes_turisticos_restriccion_11`: `CHECK ((version > 0))`
- `paquetes_turisticos_restriccion_2`: `CHECK (((horas_aviso_cancelacion >= 0) AND (horas_aviso_cancelacion <= 8760)))`
- `paquetes_turisticos_restriccion_3`: `CHECK ((((duracion_noches >= 0) AND (duracion_noches <= 29)) AND (duracion_noches < duracion_dias)))`
- `paquetes_turisticos_restriccion_4`: `CHECK ((permite_cancelacion OR (horas_aviso_cancelacion = 0)))`
- `paquetes_turisticos_restriccion_5`: `CHECK (((dificultad)::text = ANY ((ARRAY[''::character varying, 'easy'::character varying, 'moderate'::character varying, 'demanding'::character varying])::text[])))`
- `paquetes_turisticos_restriccion_6`: `CHECK (((duracion_dias >= 1) AND (duracion_dias <= 30)))`
- `paquetes_turisticos_restriccion_7`: `CHECK (((recargo_extranjero_centavos >= 0) AND (recargo_extranjero_centavos <= 100000000)))`
- `paquetes_turisticos_restriccion_8`: `UNIQUE (id, agencia_id)`
- `paquetes_turisticos_restriccion_9`: `CHECK (((precio_nacional_centavos >= 0) AND (precio_nacional_centavos <= 100000000)))`

## programaciones_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('programaciones_paquete_id_secuencia'::regclass) | No |
| paquete_id | bigint | No | Sin valor | No |
| tipo_frecuencia | character varying(24) | No | Sin valor | No |
| vigente_desde | date | No | Sin valor | No |
| vigente_hasta | date | No | Sin valor | No |
| dias_semana | jsonb | No | '[]'::jsonb | No |
| hora_salida | time without time zone | No | Sin valor | No |
| hora_encuentro | time without time zone | No | Sin valor | No |
| cupo_minimo_predeterminado | integer | No | Sin valor | No |
| cupo_maximo_predeterminado | integer | No | Sin valor | No |
| horas_cierre_compras | integer | No | Sin valor | No |
| dias_maxima_anticipacion | integer | No | Sin valor | No |
| punto_encuentro_predeterminado | text | No | ''::text | No |
| instrucciones_predeterminadas | text | No | ''::text | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `programaciones_paquete_restriccion_1`: `CHECK (((horas_cierre_compras >= 1) AND (horas_cierre_compras <= 720)))`
- `programaciones_paquete_restriccion_10`: `UNIQUE (paquete_id)`
- `programaciones_paquete_restriccion_11`: `PRIMARY KEY (id)`
- `programaciones_paquete_restriccion_2`: `CHECK ((((cupo_maximo_predeterminado >= 1) AND (cupo_maximo_predeterminado <= 500)) AND (cupo_maximo_predeterminado >= cupo_minimo_predeterminado)))`
- `programaciones_paquete_restriccion_3`: `CHECK ((vigente_hasta >= vigente_desde))`
- `programaciones_paquete_restriccion_4`: `CHECK ((vigente_hasta <= ((vigente_desde + '1 year'::interval))::date))`
- `programaciones_paquete_restriccion_5`: `CHECK ((horas_cierre_compras < (dias_maxima_anticipacion * 24)))`
- `programaciones_paquete_restriccion_6`: `CHECK (((cupo_minimo_predeterminado >= 1) AND (cupo_minimo_predeterminado <= 500)))`
- `programaciones_paquete_restriccion_7`: `CHECK (((tipo_frecuencia)::text = ANY ((ARRAY['single'::character varying, 'daily'::character varying, 'specific_weekdays'::character varying])::text[])))`
- `programaciones_paquete_restriccion_8`: `CHECK (((dias_maxima_anticipacion >= 1) AND (dias_maxima_anticipacion <= 365)))`
- `programaciones_paquete_restriccion_9`: `FOREIGN KEY (paquete_id) REFERENCES paquetes_turisticos(id) ON DELETE CASCADE`

## registros_pendientes

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| correo | character varying(320) | No | Sin valor | Sí |
| hash_token | character varying(64) | No | Sin valor | Sí |
| hash_contrasena | character varying(255) | No | Sin valor | Sí |
| nombres | character varying(100) | No | Sin valor | No |
| apellidos | character varying(100) | No | Sin valor | No |
| telefono | character varying(30) | No | ''::character varying | Sí |
| numero_documento | character varying(40) | No | ''::character varying | Sí |
| nacionalidad | character varying(80) | No | ''::character varying | No |
| creado_en | timestamp with time zone | No | Sin valor | No |
| vence_en | timestamp with time zone | No | Sin valor | No |

Restricciones:

- `registros_pendientes_restriccion_1`: `PRIMARY KEY (correo)`
- `registros_pendientes_restriccion_2`: `UNIQUE (hash_token)`

## salidas_paquete

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('salidas_paquete_id_secuencia'::regclass) | No |
| paquete_id | bigint | No | Sin valor | No |
| programacion_id | bigint | Sí | Sin valor | No |
| salida_en | timestamp with time zone | No | Sin valor | No |
| encuentro_en | timestamp with time zone | No | Sin valor | No |
| compras_abren_en | timestamp with time zone | No | Sin valor | No |
| compras_cierran_en | timestamp with time zone | No | Sin valor | No |
| cupo_minimo | integer | No | Sin valor | No |
| cupo_maximo | integer | No | Sin valor | No |
| cupos_reservados | integer | No | 0 | No |
| cupos_confirmados | integer | No | 0 | No |
| punto_encuentro | text | No | ''::text | No |
| instrucciones | text | No | ''::text | No |
| estado | character varying(20) | No | 'open'::character varying | No |
| motivo_cancelacion | text | No | ''::text | No |
| version | integer | No | 1 | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |
| es_excepcion | boolean | No | false | No |
| usuario_modificacion_id | bigint | Sí | Sin valor | No |
| cancelado_en | timestamp with time zone | Sí | Sin valor | No |
| revision_minimo_en | timestamp with time zone | Sí | Sin valor | No |
| minimo_revisado_en | timestamp with time zone | Sí | Sin valor | No |
| usuario_revision_minimo_id | bigint | Sí | Sin valor | No |

Restricciones:

- `salidas_paquete_restriccion_1`: `CHECK ((((estado)::text <> 'cancelled'::text) OR (cancelado_en IS NOT NULL)))`
- `salidas_paquete_restriccion_10`: `CHECK (((cupo_minimo >= 1) AND (cupo_minimo <= 500)))`
- `salidas_paquete_restriccion_11`: `CHECK ((((estado)::text <> 'minimum_review'::text) OR (revision_minimo_en IS NOT NULL)))`
- `salidas_paquete_restriccion_12`: `FOREIGN KEY (usuario_revision_minimo_id) REFERENCES usuarios(id) ON DELETE SET NULL`
- `salidas_paquete_restriccion_13`: `FOREIGN KEY (usuario_modificacion_id) REFERENCES usuarios(id) ON DELETE SET NULL`
- `salidas_paquete_restriccion_14`: `FOREIGN KEY (paquete_id) REFERENCES paquetes_turisticos(id) ON DELETE CASCADE`
- `salidas_paquete_restriccion_15`: `UNIQUE (paquete_id, salida_en)`
- `salidas_paquete_restriccion_16`: `PRIMARY KEY (id)`
- `salidas_paquete_restriccion_17`: `FOREIGN KEY (programacion_id) REFERENCES programaciones_paquete(id) ON DELETE SET NULL`
- `salidas_paquete_restriccion_18`: `CHECK (((estado)::text = ANY ((ARRAY['draft'::character varying, 'open'::character varying, 'confirmed'::character varying, 'closed'::character varying, 'minimum_review'::character varying, 'cancelled'::character varying, 'completed'::character varying])::text[])))`
- `salidas_paquete_restriccion_19`: `CHECK ((version > 0))`
- `salidas_paquete_restriccion_2`: `CHECK ((((estado)::text <> 'cancelled'::text) OR (length(TRIM(BOTH FROM motivo_cancelacion)) > 0)))`
- `salidas_paquete_restriccion_3`: `CHECK ((((cupo_maximo >= 1) AND (cupo_maximo <= 500)) AND (cupo_maximo >= cupo_minimo)))`
- `salidas_paquete_restriccion_4`: `CHECK ((encuentro_en <= salida_en))`
- `salidas_paquete_restriccion_5`: `CHECK ((compras_abren_en < compras_cierran_en))`
- `salidas_paquete_restriccion_6`: `CHECK ((compras_cierran_en < salida_en))`
- `salidas_paquete_restriccion_7`: `CHECK (((cupos_reservados + cupos_confirmados) <= cupo_maximo))`
- `salidas_paquete_restriccion_8`: `CHECK ((cupos_confirmados >= 0))`
- `salidas_paquete_restriccion_9`: `CHECK ((cupos_reservados >= 0))`

## sesiones_autenticacion

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | character varying(64) | No | Sin valor | No |
| usuario_id | bigint | No | Sin valor | No |
| hash_token | character varying(64) | No | Sin valor | Sí |
| ultima_actividad_en | timestamp with time zone | No | now() | No |
| vence_en | timestamp with time zone | No | Sin valor | No |
| agente_usuario | character varying(512) | No | ''::character varying | Sí |
| direccion_ip | character varying(64) | No | ''::character varying | Sí |
| revocado_en | timestamp with time zone | Sí | Sin valor | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `sesiones_autenticacion_restriccion_1`: `PRIMARY KEY (id)`
- `sesiones_autenticacion_restriccion_2`: `UNIQUE (hash_token)`
- `sesiones_autenticacion_restriccion_3`: `FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE`

## solicitudes_compra

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| turista_id | bigint | No | Sin valor | No |
| clave_solicitud | character varying(128) | No | Sin valor | No |
| hash_contenido | character varying(64) | No | Sin valor | Sí |
| compra_id | bigint | No | Sin valor | No |
| creado_en | timestamp with time zone | No | now() | No |

Restricciones:

- `solicitudes_compra_restriccion_1`: `PRIMARY KEY (turista_id, clave_solicitud)`
- `solicitudes_compra_restriccion_2`: `FOREIGN KEY (compra_id) REFERENCES compras_paquete(id) ON DELETE RESTRICT`
- `solicitudes_compra_restriccion_3`: `FOREIGN KEY (turista_id) REFERENCES usuarios(id) ON DELETE RESTRICT`

## subcategorias_atracciones

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | Sin valor | No |
| categoria_id | bigint | No | Sin valor | No |
| nombre | character varying(100) | No | Sin valor | No |
| posicion | integer | No | Sin valor | No |

Restricciones:

- `subcategorias_atracciones_restriccion_1`: `FOREIGN KEY (categoria_id) REFERENCES categorias_atracciones(id)`
- `subcategorias_atracciones_restriccion_2`: `UNIQUE (categoria_id, nombre)`
- `subcategorias_atracciones_restriccion_3`: `PRIMARY KEY (id)`

## usuarios

| Campo | Tipo/longitud | Admite NULL | Valor predeterminado | Sensible |
|---|---|---|---|---|
| id | bigint | No | nextval('usuarios_id_secuencia'::regclass) | No |
| correo | character varying(320) | No | Sin valor | Sí |
| hash_contrasena | character varying(255) | No | ''::character varying | Sí |
| nombres | character varying(100) | No | Sin valor | No |
| apellidos | character varying(100) | No | Sin valor | No |
| rol | character varying(20) | No | 'turista'::character varying | No |
| estado | character varying(20) | No | 'active'::character varying | No |
| intentos_fallidos_acceso | integer | No | 0 | No |
| bloqueado_hasta | timestamp with time zone | Sí | Sin valor | No |
| ultimo_acceso_en | timestamp with time zone | Sí | Sin valor | No |
| creado_en | timestamp with time zone | No | now() | No |
| actualizado_en | timestamp with time zone | No | now() | No |
| telefono | character varying(30) | No | ''::character varying | Sí |
| numero_documento | character varying(40) | No | ''::character varying | Sí |
| nacionalidad | character varying(80) | No | ''::character varying | No |
| identificador_google | character varying(255) | Sí | Sin valor | Sí |
| correo_verificado_en | timestamp with time zone | Sí | Sin valor | Sí |

Restricciones:

- `usuarios_restriccion_1`: `UNIQUE (correo)`
- `usuarios_restriccion_2`: `CHECK (((correo)::text = lower((correo)::text)))`
- `usuarios_restriccion_3`: `CHECK ((intentos_fallidos_acceso >= 0))`
- `usuarios_restriccion_4`: `UNIQUE (identificador_google)`
- `usuarios_restriccion_5`: `PRIMARY KEY (id)`
- `usuarios_restriccion_6`: `CHECK (((rol)::text = ANY ((ARRAY['admin'::character varying, 'turista'::character varying, 'encargado_agencia'::character varying, 'encargado_atraccion'::character varying])::text[])))`
- `usuarios_restriccion_7`: `CHECK (((estado)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying])::text[])))`
