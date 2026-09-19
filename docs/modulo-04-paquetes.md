# Módulo 4: paquetes turísticos

## Primera entrega: plantilla e itinerario

- El encargado de agencia crea y edita únicamente los paquetes de su agencia.
- Un paquete puede guardarse como borrador con nombre y duración. Para publicarlo requiere descripción de al menos 20 caracteres, precio nacional, una fotografía y al menos un día de itinerario.
- Tarifas: precio nacional por persona y costo adicional fijo por cada viajero extranjero. Los importes se almacenan en centavos.
- La edad mínima de pago se toma de la configuración de la agencia; no se duplica en el paquete.
- Duración de 1 a 30 días y noches siempre menores que los días.
- Información para el viajero: incluye, no incluye y qué llevar, con hasta 20 elementos por lista.
- Política de cancelación por paquete: permite o no cancelar y anticipación mínima en horas. En esta versión el reembolso acordado es del 100 %.
- Itinerario por días con título, descripción, actividades y atracciones publicadas. Una atracción puede repetirse en días diferentes, pero no dentro del mismo día.
- Galería de hasta seis fotografías PNG/JPEG de 5 MB; el servidor valida el contenido y lo normaliza a JPEG.
- Nginx admite hasta 42 MB en las rutas de gestión de paquetes para transportar seis imágenes codificadas en base64.
- Protección por versión para evitar que dos ediciones sobrescriban cambios.
- Una agencia desactivada no puede crear ni modificar paquetes.

## API compartida con la futura aplicación móvil

Catálogo público:

- `GET /api/v1/packages/options`
- `GET /api/v1/packages`
- `GET /api/v1/packages/{id}`
- `GET /api/v1/packages/{id}/photos/{photo}`

El catálogo solo expone paquetes publicados de agencias activas y publicadas. Permite buscar por texto, departamento, dificultad, duración, precio, fecha, disponibilidad y orden. La respuesta calcula el cupo disponible, si la ventana de compra está abierta y cuántas personas faltan para alcanzar el mínimo, sin revelar cupos retenidos internamente.

Gestión de agencia:

- `GET/POST /api/v1/agency/packages`
- `GET/PUT /api/v1/agency/packages/{id}`
- `GET /api/v1/agency/packages/attractions`
- `GET /api/v1/agency/packages/{id}/photos/{photo}`
- `PATCH /api/v1/agency/packages/{id}/departures/{departure}`
- `POST /api/v1/agency/packages/{id}/departures/{departure}/cancel`

Todas estas rutas exigen una sesión con rol `encargado_agencia`. La consulta y edición siempre se limitan a la agencia asignada.

## Programación y salidas

Cada paquete funciona como plantilla y admite una programación opcional mientras está en borrador. La frecuencia puede ser salida única, diaria o días específicos de la semana. Toda programación tiene un intervalo finito de hasta doce meses y una sola hora de salida por día.

Al guardar la programación se generan salidas concretas futuras con cupos, ventana de compra, punto de encuentro e instrucciones propios. Cambiar la programación regenera únicamente salidas futuras sin cupos retenidos o confirmados. La salida única debe ser futura para publicar; las frecuencias recurrentes se publican con una regla válida sin exigir una salida creada manualmente.

El encargado puede ajustar individualmente una salida futura: hora y punto de encuentro, instrucciones, cupos y cierre de compras. La salida queda marcada como excepción y una regeneración posterior no sobrescribe esos valores. También puede cancelarla con un motivo obligatorio; se conserva la salida, el responsable, la fecha y la razón para mantener trazabilidad. La cancelación mueve las compras afectadas a reembolso pendiente y libera sus cupos retenidos o confirmados.

La máxima anticipación, entre 1 y 365 días, determina cuándo se abre la compra. El cierre se configura en horas antes de la salida. El módulo de compras solo acepta el identificador de una salida y verifica que la hora actual esté entre ambos límites.

## Experiencia del turista

- Catálogo visual en `/packages`, disponible también antes de iniciar sesión.
- Filtros de destino, dificultad, duración, precio, fecha y venta abierta.
- Ficha en `/packages/{id}` con galería, agencia, tarifas, política de menores y cancelación, incluidos, exclusiones, recomendaciones e itinerario enlazado con las atracciones.
- Próximas salidas con cupos disponibles, hora y punto de encuentro, cierre de compra y avance hacia el cupo mínimo.
- Acceso directo desde la navegación y el inicio del turista.

El catálogo utiliza `departure_id` como unidad de disponibilidad y enlaza al flujo de compra cuando el turista inicia sesión. El comprobante, la revisión de pago y la corrección se describen en [Módulo 5](modulo-05-compras.md).

## Verificación

- Pruebas unitarias de requisitos de publicación, política de cancelación y repetición de atracciones entre días.
- Prueba de integración con PostgreSQL para permisos, aislamiento entre agencias, borrador, publicación, fotografía, itinerario, ajuste y cancelación de salidas y conflictos de versión.
- Prueba de concurrencia sincronizada: dos cambios sobre la misma versión producen una actualización exitosa y un conflicto `409`.
- Pruebas de catálogo público para agencias no publicadas o inactivas, estados ocultos, filtros inválidos, fotografías, atracciones retiradas y cálculo de disponibilidad.
- Navegación por rol y presentación de importes, duración y estados cubiertas por pruebas frontend.
- `go test ./...`, `go vet ./...`, 15 pruebas frontend, comprobación de tipos y compilación de producción completadas correctamente.
- Auditoría funcional registrada en `.medusa-auditor/results/packages-report.md`, con 22 casos aprobados y los riesgos todavía fuera del alcance de este módulo.
