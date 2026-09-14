# Módulo 2: agencias y configuración

## Alcance entregado

- Administración de agencias: creación, búsqueda, filtros, paginación y edición.
- Un encargado por agencia y una agencia por encargado, protegido también en la base de datos.
- Solo el administrador crea agencias, cambia su encargado y activa o desactiva la agencia.
- El encargado administra los datos, publicación y configuración de su propia agencia desde `/agency`.
- Edad mínima de pago configurable por agencia, de 0 a 18 años, inicialmente 6. Bajo esa edad los menores no pagan ni descuentan cupo; sus edades se registrarán en compras.
- Configuración de QR PNG/JPEG de hasta 500 KB, transferencia bancaria e instrucciones de pago.
- Es posible guardar la agencia sin métodos de pago mientras se completa su configuración.
- Una agencia desactivada queda oculta y su encargado solo puede consultar sus datos.
- No se puede cambiar el rol de un encargado con agencia asignada: primero debe reasignarse.
- Protección frente a ediciones simultáneas mediante versión del registro y aviso de conflicto.

La visibilidad queda configurada para el catálogo posterior. Este módulo no procesa pagos ni incluye todavía paquetes, compras o catálogo público de agencias.

## Cómo revisarlo

1. Ingresar como administrador en `http://127.0.0.1:3000/login`.
2. En **Usuarios**, crear una cuenta con rol **Encargado de agencia**.
3. Abrir **Agencias → Nueva agencia** y asignar esa cuenta.
4. Completar contacto, ubicación, edad mínima y los medios de pago que correspondan.
5. Guardar y comprobar la agencia en el listado.
6. Ingresar con la cuenta del encargado y abrir **Mi agencia** para configurar la edad mínima y los medios de pago.

No se cargaron agencias de ejemplo; los datos reales los incorpora el usuario.

## Verificación técnica

- Migración `000003_agencies.sql` aplicada en desarrollo.
- Pruebas Go con PostgreSQL en esquemas temporales: permisos por rol, CSRF, aislamiento entre agencias, reasignación, edición concurrente y asignación simultánea del mismo encargado.
- Validaciones de edad, datos bancarios y contenido real de imágenes QR.
- Suite completa de Go y `go vet ./...` aprobados.
- Frontend: comprobación de tipos, pruebas y compilación de producción aprobadas.

Se mantiene la configuración de ejecución local documentada en `modulo-01-usuarios.md`.
