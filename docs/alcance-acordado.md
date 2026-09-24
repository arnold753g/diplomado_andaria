# Andaria para el diplomado — decisiones acordadas

Entrega por módulos revisables. Base Go, Nuxt/Vue y PostgreSQL; conservar estilo visual, reutilizar logo y posponer cambio de paleta. Tarija inicialmente, con geografía ampliable a Bolivia. Datos turísticos cargados por el usuario. Web publicada en Internet y posterior app Flutter compartiendo backend.

1. Usuarios, roles, perfil y acceso por correo o Google.
2. Agencias, encargado y configuración de tarifas/pagos.
3. Atracciones, asignaciones, catálogo y favoritos.
4. Paquetes, itinerarios, salidas y cupos.
5. Compras, comprobantes y validación.
6. Cancelaciones y reembolsos.
7. Home final, pruebas integrales y publicación.

## Responsabilidades

- Solo el administrador crea agencias, crea atracciones y asigna encargados y roles.
- Un encargado de agencia gestiona una agencia; inicialmente cada agencia tiene un encargado.
- Un encargado de atracción gestiona una o varias atracciones; cada atracción tiene un encargado.
- Encargados publican directamente su contenido. La agencia crea sus paquetes.
- Turistas registran favoritos de atracciones y compran varios cupos a su nombre.

## Reglas comerciales para módulos posteriores

- Cada paquete es una plantilla. La agencia configura una salida única, salidas diarias o días específicos de la semana dentro de un intervalo máximo de doce meses; el sistema genera las salidas concretas.
- Cada salida conserva fecha, hora, cupo mínimo, cupo máximo, cierre de compra, máxima anticipación de compra, punto de encuentro e instrucciones.
- El itinerario se organiza por días con título, descripción, actividades y atracciones opcionales. Una atracción puede aparecer en varios días.
- Para publicar un paquete se exige nombre, descripción, precio, duración, una fotografía y al menos un día de itinerario.
- El punto de encuentro y las instrucciones se definen en cada salida porque pueden cambiar según la fecha.
- La compra siempre referencia una salida generada. Nunca acepta una fecha libre, pasada ni fuera de la ventana de compra de esa salida.
- Edad mínima de pago configurable por agencia, inicialmente 6 años.
- Menores bajo esa edad no pagan ni descuentan cupo; se registran edades, sin exigir nombres de acompañantes.
- Desde la edad mínima, pagan y descuentan cupo. Separar total de viajeros de cupos ocupados.
- Tarifa base y costo adicional por extranjero. Modelar cantidades para grupos mixtos sin aplicar nacionalidad del comprador a todo el grupo.
- Pago del 100 % por QR o transferencia con comprobante, validado por el encargado de agencia.
- Acumular compras para alcanzar el mínimo de salida; informar esa condición. Si no se alcanza antes del límite, cancelar y tramitar devolución.
- Paquete define si permite cancelación y con cuánto tiempo de anticipación; devoluciones del 100 % en esta versión.
- Turista puede subir QR para solicitar reembolso. Agencia registra devolución y comprobante. Subir QR no significa que el dinero haya sido devuelto.
- Cancelar una salida debe registrar todas las compras afectadas y dar seguimiento a reembolsos.

## Decisiones cerradas para compras y reembolsos

- No existe una reserva previa durante la transferencia. El cupo se retiene de forma atómica cuando el servidor acepta el comprobante.
- Al cerrar las compras sin alcanzar el mínimo, la salida queda pendiente de decisión. La agencia debe confirmar antes de cancelarla y enviar las compras a reembolso.
- Una cancelación válida del turista libera cupos inmediatamente y no requiere aprobación de la agencia.
- El turista puede proporcionar un QR o datos bancarios para la devolución.
- La agencia dispone de 72 horas desde el inicio del reembolso y adjunta un comprobante de hasta 5 MB para marcarlo como completado.

## Decisiones cerradas para home y paneles

- `/` es la portada pública. Las áreas `/app`, `/agency` y `/admin` continúan protegidas por sesión y rol.
- La portada usa contenido publicado y disponible: próxima salida comprable, paquetes con venta abierta y atracciones recientes. No requiere una administración manual de destacados.
- El mensaje principal es «Explora Bolivia, comenzando por Tarija», conservando la posibilidad de ampliar el catálogo a otros departamentos.
- El turista ve su próxima compra, compras activas, favoritos y reembolsos.
- La agencia ve pagos por revisar, salidas que requieren decisión por cupo mínimo y reembolsos próximos a vencer o atrasados.
- El encargado de atracciones ve totales de asignadas, publicadas, borradores e inactivas.
- El administrador ve cifras globales de usuarios, agencias, atracciones, paquetes y compras.
- Todos los contadores operativos se calculan en el backend con el alcance del usuario autenticado; el frontend no reconstruye permisos ni mezcla datos de otros encargados.
