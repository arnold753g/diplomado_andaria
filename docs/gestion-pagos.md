# Gestionar pagos por paquete y salida

Implementación del 27/10/2026, actualizada el 28/10/2026. La vista del encargado de agencia está en `/agency/purchases`, con el nombre **Gestionar pagos** en el menú, título y accesos del panel.

## Organización

- Vista general agrupada por paquete y por salida concreta (fecha y hora de Bolivia).
- Vista inicial **Habilitadas**: solo salidas futuras con autorización vigente, agencia activa y mínimo confirmado. Las salidas sin autorización están en **Pendientes**; las pasadas, completadas y canceladas en **Historial**. **Reembolsos** reúne las salidas con devoluciones pendientes o realizadas, incluso si ya pasaron o fueron canceladas. Una misma salida puede aparecer en Reembolsos e Historial.
- Indicadores calculados en PostgreSQL antes de paginar. Los avisos de pagos en revisión, salidas listas para autorizar y reembolsos pendientes/vencidos consideran todas las salidas que coinciden con nombre, paquete y fechas, independientemente de la vista activa. Los contadores de navegación cuentan salidas; el aviso de reembolsos cuenta compras. El importe confirmado corresponde a la vista activa; no es un saldo bancario ni una conciliación automática.
- Filtros por nombre, paquete e intervalo de fechas. El intervalo incluye los dos días seleccionados en America/La_Paz. La URL conserva la vista con `view=authorized|pending|refunds|history`.
- Se retiraron los accesos **Ver programación** de esta pantalla. El indicador de reembolsos del panel de agencia abre directamente la vista Reembolsos.
- Prioridad para salidas con pagos, correcciones, reembolsos o decisiones pendientes; la paginación no altera los totales.
- Detalle por salida con pestañas Pagos, Turistas y Reembolsos. Los comprobantes se abren solo cuando se consulta una compra; las acciones existentes de revisión y devolución se conservan.
- Lista paginada de compradores responsables con teléfono, correo y cantidades de adultos y menores. No se piden nombres o documentos nuevos de los acompañantes.
- Lista imprimible completa de compras confirmadas, protegida por sesión y agencia, con escape HTML y `Cache-Control: private, no-store`. Puede imprimirse o guardarse como PDF desde el navegador. No es una identificación individual de todos los acompañantes.

## Autorización manual

El estado comercial `confirmed` y el cupo mínimo alcanzado no autorizan el viaje. Se mantiene ese estado para compatibilidad del contrato web/Android. La interfaz de gestión distingue Mínimo pendiente, Lista para confirmar, Salida autorizada, Decisión por mínimo, Cancelada y Salida pasada.

El encargado puede autorizar solamente una salida futura de su agencia activa, con el mínimo de cupos confirmados y estado comercial habilitado. La API verifica estos requisitos y la versión dentro de una transacción con bloqueo de filas. La segunda confirmación concurrente sobre la misma versión devuelve 409; una salida ya autorizada no admite otra autorización sobre su versión actual.

Se registran la fecha y el usuario de la última autorización. Si antes de viajar los cupos confirmados quedan por debajo del mínimo, o si la salida se cancela, un disparador de PostgreSQL retira su vigencia y registra fecha/motivo. No se cancela toda la salida ni se reembolsa automáticamente a los demás compradores por perder el mínimo. Recuperar el mínimo exige una nueva autorización explícita.

Se conserva la última decisión, no un historial inmutable de todas las autorizaciones. La autorización retirada conserva su responsable/fecha hasta la siguiente confirmación. Las salidas anteriores a esta entrega quedan sin autorización operativa registrada; no se atribuye una decisión a un encargado sin evidencia.

Los menores exentos siguen sin consumir cupo, conforme a la política vigente. Se muestran por separado turistas totales y cupos computables para el mínimo.

## Cancelación y devoluciones

En el detalle de una salida futura de la agencia activa, **Cancelar salida** permite cancelar aunque ya estuviera autorizada. El encargado selecciona **Falta de cupos** o **Motivos externos**, explica el motivo (5–900 caracteres) y confirma. Se registra la causa y explicación en el motivo de cancelación, junto con responsable y fecha; la salida no se puede cancelar dos veces ni después de empezar.

Se reutiliza la cancelación transaccional existente: bloquea la salida, comprueba la versión, retira la autorización, libera los cupos y pasa las compras confirmadas, en revisión o en corrección a `refund_pending`. Las compras que ya estaban pendientes de devolución mantienen su plazo anterior; las nuevas tienen el plazo vigente de 72 horas. Las ya reembolsadas no cambian. Cancelar una salida sin compras no crea devoluciones. No se ejecuta una transferencia bancaria automática.

La pantalla abre Reembolsos al terminar la cancelación. Para completar cada devolución:

1. El turista registra su QR o cuenta bancaria en la compra.
2. La agencia realiza la devolución por el medio indicado.
3. En **Ver detalle**, la agencia adjunta el comprobante y pulsa **Marcar devolución completada**. Sin destino y comprobante no se puede cerrar el reembolso.

Las salidas canceladas quedan accesibles en Historial y, cuando tienen devoluciones, también en Reembolsos. Sus avisos siguen visibles aunque ya no aparezcan en Habilitadas. Se conserva el flujo existente de la web y APK para indicar el destino de devolución; este cambio no requiere una nueva versión de la APK. No se incorporaron notificaciones por correo: la configuración SMTP sigue pendiente.

## API

Todas estas rutas exigen sesión con rol encargado_agencia y filtran por la agencia asignada:

- `GET /api/v1/agency/payments/departures`: resumen, paquetes disponibles en el filtro e indicadores completos. Parámetros `page`, `limit`, `search`, `package_id`, `from`, `to`, `state`. Incorpora `pending`, `refunds` e `history`; conserva los filtros anteriores. La web solicita `authorized` al entrar, mientras que omitir `state` sigue devolviendo todas las salidas por compatibilidad. `summary` contiene alertas y contadores antes del filtro de estado; `totals` describe el conjunto seleccionado.
- `GET /api/v1/agency/payments/departures/{departure}`: detalle operativo y financiero de la salida.
- `POST /api/v1/agency/payments/departures/{departure}/authorize`: autorización manual con `{version}` y protección CSRF.
- `GET /api/v1/agency/payments/departures/{departure}/roster`: documento HTML imprimible de todas las compras confirmadas.
- `POST /api/v1/agency/packages/{id}/departures/{departure}/cancel`: cancelación existente con `{version, reason}`, validación de agencia y protección CSRF. El detalle incorpora `can_cancel` para mostrar la acción cuando procede.
- `GET /api/v1/agency/purchases`: incorpora filtros `departure_id`, `package_id` y el estado `cancelled`, conservando su respuesta y paginación previas.

Las rutas existentes de comprobantes, revisión, cancelación por mínimo y devolución permanecen vigentes. La lista impresa usa una sesión del navegador en el mismo sitio que la API, como el despliegue actual.

## Base de datos y despliegue

La migración **16**, `000016_autorizacion_salidas.sql`, es aditiva sobre el esquema español. Agrega a `salidas_paquete`:

| Columna | Finalidad |
|---|---|
| `autorizado_en` | Fecha de la última autorización |
| `usuario_autorizacion_id` | Responsable de esa autorización |
| `autorizacion_revocada_en` | Momento en que pierde vigencia |
| `motivo_revocacion_autorizacion` | Motivo del retiro |

Añade restricciones de coherencia, un índice y el disparador `validar_autorizacion_salida()`. La API comprueba la versión 16 al arrancar. Los modelos y claves públicas existentes mantienen compatibilidad; no cambia el formulario de compra de la APK.

La reorganización y cancelación desde Gestionar pagos del 28/10 no agregan migraciones; requieren reconstruir backend y frontend, manteniendo el esquema en versión 16.

La base local ya se migró con respaldo privado previo y conteos conservados. La API local se reconstruyó para la vista nueva. **No se modificó el VPS.** Para publicar, preparar los roles limitados y la configuración de seguridad documentados, reconstruir backend/migrador/frontend, detener las aplicaciones, ejecutar `migrate` y levantar los servicios solo si termina correctamente. Validar Nginx y los recorridos públicos/privados. Ver [procedimiento coordinado](base-datos-espanol.md).

## Verificaciones

- Suite Go completa con PostgreSQL local, instalación vacía y actualización desde el esquema anterior; `go vet`.
- Pruebas nuevas con dos agencias y roles turista/admin/atracción: acceso 401/403/404 según corresponda, filtros, totales sobre varias páginas y 25 compras confirmadas.
- Confirmación concurrente: una autorización y un conflicto; autorización sin mínimo rechazada, agencia inactiva rechazada, cancelación real de una compra revoca la autorización y recuperar el mínimo no la reactiva.
- HTML imprimible completo y escapado, sin pagos no confirmados ni cacheado de información personal.
- 30 pruebas frontend, comprobación de tipos y compilación de producción Nuxt en una copia aislada. Verificación visual con datos ficticios en escritorio y 390 px, sin desbordamiento horizontal.
- Filtros Habilitadas/Pendientes/Reembolsos/Historial, salidas pasadas, contadores globales y aislamiento entre agencias. Cancelación concurrente de una salida autorizada: un éxito y un conflicto, autorización revocada, 28 compras pendientes de devolución, plazos anteriores conservados y destino de devolución visible para el turista. Se rechazan cancelaciones de otra agencia, pasadas o sin motivo suficiente.
- Recorrido adicional: motivo obligatorio, cancelación por causa externa, acceso automático a Reembolsos, bloqueo de cierre sin QR/cuenta y acceso a la cancelada desde Historial. Sin errores ni advertencias en consola. Capturas ficticias del 28/10 en [evidencias](evidencias/gestion-pagos-2026-10-28/).
- Recorrido en navegador: acceso de agencia, vista por paquete/salida, autorización con responsable/fecha, pestaña Turistas, segunda página y lista imprimible. Consola sin errores ni advertencias en ese recorrido.
- Capturas en [evidencias](evidencias/gestion-pagos-2026-10-27/), exclusivamente con datos ficticios. No se utilizaron ni enviaron correos reales ni movimientos bancarios.
