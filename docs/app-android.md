# App Android de Andaria

## Decisiones

Flutter + Dart en `mobile/`, versionado junto al backend y la web. Solo turistas. Catálogo público, cuentas y datos compartidos, español, bolivianos y horarios Bolivia. Tema claro con superficies blancas/grises, texto oscuro, fotografías naturales y botones verde lima `#BAFE06`.

## Contrato móvil añadido

### Autenticación

`GET /api/v1/auth/mobile/options` devuelve `registration_enabled`, `google_enabled` y el identificador OAuth público `google_server_client_id`.

`POST /api/v1/auth/google/mobile` recibe `{ "id_token": "..." }`. Se verifican firma, audiencia, emisor, expiración, identidad y correo verificado. Reutiliza la identidad Google existente; no vincula automáticamente una cuenta del mismo correo. Solo emite sesiones para turistas activos. Mantiene el modelo de cookies revocables y CSRF de la web.

La configuración de clientes Android y firmas está descrita en `mobile/README.md`. Las pruebas automatizadas validan rechazos y resolución de identidad; no sustituyen una prueba con una cuenta Google real.

### Compras

`POST /api/v1/me/purchases` admite ahora:

- Cabecera opcional `Idempotency-Key`: de 16 a 128 caracteres alfanuméricos, guion o guion bajo.
- Campo opcional `expected_total_cents`: total positivo que el cliente mostró al usuario.

Una compra nueva responde `201`. Un reintento idéntico del mismo turista devuelve la compra existente con `200`, sin volver a retener cupos. Reutilizar una clave con datos distintos produce `409 PURCHASE_REQUEST_CONFLICT`. Un precio vigente distinto produce `409 PURCHASE_PRICE_CHANGED` y no crea la compra.

La migración `000013_purchase_idempotency.sql` añade `purchase_requests`. Un bloqueo transaccional por usuario/clave serializa los reintentos; compra, cupos y clave se guardan en la misma transacción. El historial de claves aceptadas se conserva para proteger reintentos tardíos. Las solicitudes web que omiten los campos nuevos mantienen su comportamiento.

### Contacto comercial

Las opciones de pago y las respuestas de compras incluyen los campos opcionales `agency_phone` y `agency_email`, obtenidos del registro comercial de la agencia. No se utiliza el correo del administrador como alternativa. La consulta de compras conserva su restricción por propietario; no se añade un endpoint público.

La app muestra los datos seleccionables y permite abrir el teléfono o preparar un correo con la referencia de compra. No envía correos ni inicia llamadas automáticamente. Si el teléfono contiene una extensión u otro formato ambiguo, mantiene el dato para copiarlo y omite el botón de marcación. Si no hay una aplicación disponible, muestra una alternativa para copiar los datos. Los clientes que omiten o desconocen estos campos mantienen su funcionamiento.

## Validación ejecutada

- Análisis Flutter sin incidencias y 31 pruebas aprobadas: sesión, CSRF, expiración, rol turista, respuestas tardías, aislamiento entre cuentas, imágenes, precios por grupo, navegación con texto ampliado, rechazo de un precio desactualizado y apertura de contactos con correos codificados correctamente.
- Backend: `go vet ./...` y suite completa `go test ./... -count=1` aprobados con PostgreSQL. Se verificaron reintentos simultáneos, separación entre turistas, cambio de precio, compatibilidad web y validación del acceso Google.
- Web: 16 pruebas y revisión de tipos aprobadas.
- Android: recorrido público de catálogo, detalle, acceso y ampliación de fotografías aprobado en emulador. Recorrido completo aprobado: acceso, edición de perfil, favoritos, viajeros con distintas tarifas, comprobante, corrección solicitada por agencia, confirmación, cancelación, destino y comprobante de devolución, y cierre de sesión. Se utilizaron comprobantes simulados, peticiones reales al backend y un esquema temporal eliminado al terminar.
- Android: registro nuevo, rechazo de contraseña actual incorrecta, cambio de contraseña, revocación de dos sesiones y acceso con la nueva contraseña aprobados contra el servidor aislado. La prueba simula la entrada de texto.
- Pendiente de validación externa: cuenta Google real con cliente Android configurado, teclado y selector de fotografías del sistema en un teléfono físico y ejecución remota de GitHub Actions.

## Revisión antes del pago

La app muestra los datos del comprador y permite revisarlos antes de enviar el comprobante. Si el servidor devuelve `PURCHASE_PRICE_CHANGED`, deshabilita el envío con la tarifa anterior y ofrece regresar al detalle para consultar el precio vigente. Si la salida se cierra o pierde sus cupos, bloquea el envío con la disponibilidad anterior y ofrece revisar las salidas. Si el turista ya transfirió dinero, pide consultar a la agencia antes de realizar otra transferencia. La política de cancelación también se muestra antes del pago. Un reintento idéntico conserva su clave sin generar otra compra. La obtención de esa clave se serializa por solicitud para que dos llamadas simultáneas reciban el mismo UUID. La confirmación admite un solo diálogo a la vez y la cuenta se comprueba tanto al confirmar como después de recuperar la clave, antes de enviar la compra.

Después de una revocación confirmada por el servidor, la app oculta inmediatamente los datos privados. Un fallo al limpiar el almacenamiento local no transforma el cambio de contraseña confirmado en un error de conexión; una respuesta 401 conserva su significado y tampoco deja datos privados visibles.

## Reglas comerciales conservadas

No existe reserva previa a la transferencia. La compra aceptada retiene cupos y queda en revisión. Los menores bajo la edad de pago configurada por agencia no pagan ni ocupan cupo; la nacionalidad se registra por tipo de adulto y por menor. Las cancelaciones dependen de la política y fecha del paquete. Cancelar libera cupos y abre el seguimiento de devolución; aportar un destino no significa que la devolución esté completada.
