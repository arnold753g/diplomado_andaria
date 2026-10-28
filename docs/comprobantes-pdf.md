# Comprobantes de compra PDF

Implementación del 28/10/2026. El backend Go genera un único formato A4 para web y Android con `github.com/signintech/gopdf` 0.33.0. El logo transparente y las fuentes estáticas Outfit Regular/SemiBold se incluyen en el ejecutable con `go:embed`: el contenedor no necesita Chromium, TeX, fuentes del sistema ni servicios externos. La licencia OFL se conserva en los assets; `docs/tools/preparar-fuentes-comprobante.py` permite regenerar las instancias desde la fuente variable original con fontTools, solo en desarrollo.

## Contenido y reglas

- Referencia original, fecha de compra y fecha de generación en horario de Bolivia.
- Datos de la agencia y comprador responsable; salida, encuentro, punto de reunión y duración.
- Adultos nacionales/extranjeros, menores nacionales/extranjeros que pagan y menores exentos. Las tarifas provienen de los precios enteros en centavos guardados en la compra; no se recalculan con el catálogo vigente ni se aplica un precio promedio.
- Importe y medio de pago. El PDF no incorpora el QR bancario, la imagen del depósito ni las cuentas de reembolso.
- Estado financiero actual, fecha de confirmación conservada y situación operativa de la salida, de forma independiente. La agencia inactiva no aparece como autorizada para operar.
- Cancelaciones, plazo de devolución y referencia/fecha de reembolso cuando corresponda, condiciones contratadas y recomendaciones.
- Es una constancia interna de compra, no una factura fiscal ni una autorización de viaje. Las imágenes bancarias mantienen sus rutas privadas existentes.

Disponible para compras `confirmed`, `refund_pending` y `refunded`. Los pagos en revisión/corrección no se presentan como pagados. Un reembolso iniciado antes de confirmar el pago puede tener constancia de compra/devolución, pero no se inventa su fecha de confirmación. Canceladas/rechazadas sin devolución no emiten PDF. Cada descarga muestra el estado al generar esa copia; un archivo descargado no cambia posteriormente.

## Datos históricos y migración 17

Migración aditiva `000017_comprobantes_compra.sql`, sin borrar tablas ni compras:

- `compras_paquete.datos_comprobante`: copia JSON de agencia, paquete, salida y condiciones al insertar la compra. El disparador sustituye cualquier valor de entrada y conserva la copia en actualizaciones posteriores.
- `compras_paquete.pago_confirmado_en`: conserva la primera confirmación registrada, aunque luego se solicite o complete una devolución.
- Las compras anteriores reciben los datos comerciales que existen al migrar, con `origen=migracion` y fecha de recuperación. El PDF informa de esta limitación; no atribuye nombres o condiciones históricos desconocidos. Solo se recupera la fecha de confirmación de las compras que siguen confirmadas y tienen `revisado_en`.

El modelo mantiene estas claves de persistencia en español y excluye la copia comercial del JSON público. La API exige esquema 17 al arrancar. El JSON de compra incorpora `receipt_available` y `payment_confirmed_at`.

La base local se actualizó con respaldo privado y conteos de tablas conservados. Para el VPS: respaldar, detener aplicaciones, reconstruir backend/migrador y frontend, ejecutar el migrador hasta versión 17 y levantar servicios tras su éxito. Publicar la nueva APK después de la API. No es necesario cambiar credenciales ni registrar permisos nuevos en Google para esta función.

## API privada

| Ruta | Acceso |
|---|---|
| `GET /api/v1/me/purchases/{id}/receipt` | Turista propietario |
| `GET /api/v1/agency/purchases/{id}/receipt` | Encargado de la agencia de la compra |

Respuesta `application/pdf`, `Cache-Control: private, no-store` y `nosniff`. Por defecto se abre en línea; `?download=1` solicita descarga con un nombre numérico controlado por el servidor. No incluye credenciales en las URL ni crea archivos públicos. Una compra ajena devuelve 404; un rol incorrecto, 403; falta de sesión, 401; estado financiero no disponible, 409.

## Web y Android

En Mis compras y en el detalle de revisión de la agencia aparecen Ver comprobante PDF y Descargar PDF cuando la API lo permite. La descarga usa la sesión del navegador; el servidor vuelve a verificar permisos.

Android obtiene los bytes con la sesión privada y los valida por tipo MIME y firma PDF. La descarga rechaza respuestas de una sesión anterior. La pantalla usa `printing` para vista previa, compartir e imprimir; el diálogo Android permite Guardar como PDF. El cambio de cuenta retira la vista privada. El PDF se mantiene en memoria durante la consulta; compartir o guardarlo crea una copia bajo control del usuario que no se revoca al cerrar sesión.

## Verificación

- Suite Go con PostgreSQL: instalación vacía/actualización, datos históricos, copia comercial inmutable tras cambios del catálogo, fecha de confirmación y aislamiento de turistas/agencias/roles. PDF real, cabeceras privadas, vista en línea y descarga.
- Tarifas mixtas: 4 turistas, 3 cupos, un menor exento, recargo de extranjeros y total exacto de Bs 850,00. Fuentes con tildes y ñ, textos largos y varias páginas sin pérdida de las 100 recomendaciones de prueba.
- Comprobación de tipos Nuxt y 30 pruebas frontend. Flutter: análisis y 43 pruebas, incluidas credenciales, firma/tipo de PDF y cambio de sesión durante la descarga.
- Recorrido web con turista y agencia: apertura del PDF desde ambas cuentas. Compilación Nuxt de producción correcta en una copia aislada.
- Prueba nativa aprobada en Android con un AVD desechable: descarga HTTP autenticada, vista previa, rasterización con PdfRenderer, captura y retiro de la pantalla privada al cerrar sesión. La prueba valida el marcador del servidor ficticio y se ejecuta únicamente contra su esquema temporal, eliminado al terminar.
- APK release 0.1.6+7 generada con la configuración HTTPS de producción y la firma local existente. Verificada con apksigner (firma v2 válida) y aapt (identificador y versiones correctos). Entrega: `output/apk/Andaria-0.1.6+7.apk`, con su archivo SHA-256. Esta APK necesita que el servidor de producción reciba primero la migración 17 y la nueva API para mostrar los comprobantes.
- Muestras exclusivamente ficticias; no se realizaron pagos, cancelaciones ni devoluciones reales. Ejemplo en `output/pdf/Comprobante-Andaria-ejemplo.pdf` y capturas en `docs/evidencias/comprobantes-pdf-2026-10-28/`.

El envío de recibos por correo no forma parte de esta entrega: SMTP sigue pendiente. El VPS no se modifica durante la implementación local.
