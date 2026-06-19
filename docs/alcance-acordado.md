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

- Edad mínima de pago configurable por agencia, inicialmente 6 años.
- Menores bajo esa edad no pagan ni descuentan cupo; se registran edades, sin exigir nombres de acompañantes.
- Desde la edad mínima, pagan y descuentan cupo. Separar total de viajeros de cupos ocupados.
- Tarifa base y costo adicional por extranjero. Modelar cantidades para grupos mixtos sin aplicar nacionalidad del comprador a todo el grupo.
- Pago del 100 % por QR o transferencia con comprobante, validado por el encargado de agencia.
- Acumular compras para alcanzar el mínimo de salida; informar esa condición. Si no se alcanza antes del límite, cancelar y tramitar devolución.
- Paquete define si permite cancelación y con cuánto tiempo de anticipación; devoluciones del 100 % en esta versión.
- Turista puede subir QR para solicitar reembolso. Agencia registra devolución y comprobante. Subir QR no significa que el dinero haya sido devuelto.
- Cancelar una salida debe registrar todas las compras afectadas y dar seguimiento a reembolsos.

## Propuesta pendiente de cerrar en el módulo de compras

Se propuso retener cupos durante diez minutos mientras se paga, con reloj del servidor. Comprobante dentro del plazo convierte retención a compra en revisión sin descontar dos veces. Al vencer sin comprobante se liberan cupos. Un comprobante tardío obliga a revalidar disponibilidad y, si no existe, a revisar el pago y tramitar su devolución.

El usuario solicitó evaluar una retención corta; los diez minutos son una propuesta inicial, no una regla definitivamente confirmada. Falta cerrar también el plazo de revisión de comprobantes y la fecha límite para alcanzar el mínimo. Resolver estos puntos al desarrollar compras, sin bloquear los primeros módulos.
