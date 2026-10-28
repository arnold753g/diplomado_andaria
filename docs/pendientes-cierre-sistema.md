# Pendientes para cerrar la version actual de Andaria

Revision del 27/10/2026 basada en el codigo local y los informes de implementacion. No se realizo una inspeccion nueva del VPS ni de Google Cloud. La existencia de una pantalla no equivale a una nueva prueba de su flujo completo.

## Cambio aplicado hoy

La navegacion web del encargado de agencia contiene Inicio, Mi agencia, Gestionar paquetes, Revisar compras y Mi perfil. Se retiraron Explorar atracciones y Explorar paquetes de su menu. El mismo componente sirve la navegacion de escritorio y la adaptable a movil. Los catalogos siguen siendo publicos; ocultar enlaces no agrega una restriccion de acceso.

Se actualizaron las pruebas existentes de navegacion y se comprobaron los accesos del turista y del visitante. La aplicacion Flutter actual esta orientada al turista; no se agrego un panel movil de gestion de agencias.

## Modulos presentes en el codigo

Usuarios y roles; acceso con contrasena y Google; perfil; agencias; atracciones y favoritos; paquetes e itinerarios; programaciones y salidas; compras, comprobantes, revision de pagos, cancelaciones y reembolsos; paneles por rol. El esquema en espanol y el endurecimiento de seguridad tienen informes separados. Esto describe cobertura implementada, no garantiza por si solo el cierre funcional de cada recorrido.

## Trabajo pendiente

| Prioridad | Trabajo | Criterio de cierre |
|---|---|---|
| Funcional | Recuperacion de contrasena | Implementar solicitud y restablecimiento en API, web y APK, con respuesta que no enumere cuentas, token aleatorio almacenado como hash, vencimiento, uso unico, limites y revocacion de sesiones; probar casos negativos. Actualmente existe cambio de contrasena para sesiones autenticadas, pero no recuperacion para quien la olvido. |
| Correo, al final segun la decision del usuario | Configurar SMTP real | La verificacion de registro ya tiene implementacion. Configurar credenciales privadas, comprobar entrega, reenvio y vencimiento, y verificar la recuperacion cuando exista. No pedir ni guardar credenciales en este informe. |
| Publicacion | Desplegar la version actual | Aplicar conjuntamente roles PostgreSQL limitados, migracion 15 y configuracion de seguridad; comprobar salud, Google web/Android, HTTPS, CSP y cadena de proxies. Confirmar en Google Cloud las condiciones de publicacion para usuarios externos. No se afirma que estas configuraciones del VPS hayan cambiado hoy. |
| Operacion | Respaldos y recuperacion | Definir destino externo y custodia de claves, activar respaldos cifrados, realizar una restauracion real en un entorno separado y comprobar alertas/monitores. Verificar cifrado de discos/volumenes. |
| Seguridad adicional | Cerrar decisiones pendientes | RLS no esta implementado: requiere diseno y pruebas de identidad transaccional y conexiones reutilizadas; hoy el aislamiento esta en la API. Ejecutar escaneos remotos de CI/imagenes y resolver la excepcion temporal de dependencias antes de su vencimiento. Un limitador compartido es necesario antes de multiples replicas; CAPTCHA se decide segun abuso, no se declara requisito nuevo. |
| Aceptacion | Recorrer los flujos completos | Probar con los cuatro roles: crear/publicar oferta y salida, comprar, revisar/corregir pago, cancelar y reembolsar, incluyendo propiedad, concurrencia y errores. Probar el APK exacto en un telefono fisico. Las pruebas anteriores no se presentan como ejecutadas de nuevo hoy. |
| Entrega | Cerrar documentacion y evidencias | Comprobar manual de instalacion, entrega reproducible, diccionario y diagramas actuales. El ultimo intento documentado del PDF encontro un problema de configuracion Biber/MiKTeX; no se intento compilar de nuevo hoy. Validar ademas los datos y respaldos academicos pendientes. |

Orden practico: terminar recuperacion de contrasena, cerrar pruebas funcionales por rol, configurar el correo al final como se acordo y despues realizar despliegue y aceptacion sobre la version final. La preparacion de respaldos y seguridad precede al despliegue.

## Referencias

- [Esquema en espanol y despliegue coordinado](base-datos-espanol.md).
- [Seguridad implementada y controles pendientes](seguridad-implementacion.md).
- [Criterios historicos de cierre y aceptacion](informe-correcciones-y-pendientes-2026-10-11.md).

## Entrega de Gestionar pagos (27/10/2026)

Implementados agrupamiento por paquete/salida, autorizacion manual con responsable/fecha y retiro de vigencia al perder el minimo, lista de compradores por salida, impresion completa y paginacion. La migracion 16 y la API se aplicaron localmente; publicacion en VPS pendiente. Ver [detalle](gestion-pagos.md).
