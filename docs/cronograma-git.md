# Andaria — cronograma propuesto de entregas

Fecha de preparación: 20 de septiembre de 2026. Se asume que el inicio solicitado es el **10 de junio de 2026**, con zona horaria Bolivia (UTC−04:00).

Este es un escenario de planificación para organizar la versión reducida, no un registro de actividad comprobada. En la revisión el repositorio todavía no tenía commits ni remoto: no existe evidencia de un commit o una subida el 10 de junio. La fecha de un commit y la fecha de publicación en un servidor son distintas. Este documento no modifica ninguna de ellas.

Los intervalos se distribuyen de forma irregular entre 7 y 14 días. Los títulos corresponden a Andaria (Go, Nuxt/Vue y PostgreSQL), tomando como referencia `alcance-acordado.md`; no al proyecto de pizzas del manual recibido.

| N.º | Fecha propuesta | Intervalo | Mensaje de commit propuesto | Contenido / referencia |
| --- | --- | --- | --- | --- |
| 1 | 10/06/2026 | — | `chore: inicializa el repositorio de Andaria` | README, .gitignore y plantilla de entorno sin secretos |
| 2 | 19/06/2026 | 9 días | `docs: define el alcance de Andaria para el diplomado` | Usuarios, agencias, atracciones, paquetes y compras; docs/alcance-acordado.md |
| 3 | 01/07/2026 | 12 días | `chore: configura Go, Nuxt y PostgreSQL con Docker` | Dependencias, servicios, configuración y endpoint de salud |
| 4 | 09/07/2026 | 8 días | `feat: incorpora migraciones de usuarios y sesiones` | Migración inicial, modelos y comandos migrate/seed |
| 5 | 22/07/2026 | 13 días | `feat: implementa autenticacion, Google y administracion de usuarios` | Registro, login, logout, sesiones, CSRF y límites de acceso |
| 6 | 01/08/2026 | 10 días | `feat: agrega la interfaz de acceso, perfiles y administracion` | Cuatro roles, perfil, cambio de contraseña y administración |
| 7 | 15/08/2026 | 14 días | `test: documenta y verifica el modulo de usuarios y acceso con Google` | OAuth, migración Google e interfaz de acceso |
| 8 | 22/08/2026 | 7 días | `feat: incorpora la API de agencias y asignacion de encargados` | Migración de agencias, administración y asignación de encargado |
| 9 | 02/09/2026 | 11 días | `feat: agrega el panel de agencias y configuracion de pagos` | Edad mínima, QR, transferencia y panel de la agencia |
| 10 | 14/09/2026 | 12 días | `docs: registra la entrega de usuarios y agencias y sus limites` | Pruebas de módulos 1–2 y documentación de sus reglas |
| 11 | 23/09/2026 | 9 días | `feat: agrega atracciones, catalogo y favoritos` | Módulo 3: administración, encargados y favoritos de turistas |
| 12 | 06/10/2026 | 13 días | `feat: implementa paquetes e itinerarios turisticos` | Módulo 4: paquetes creados por agencias e itinerarios |
| 13 | 14/10/2026 | 8 días | `feat: gestiona salidas y disponibilidad de cupos` | Fechas, capacidad y mínimo de participantes |
| 14 | 25/10/2026 | 11 días | `feat: incorpora compras y validacion de comprobantes` | Módulo 5: viajeros, tarifas, comprobantes y validación de pagos |
| 15 | 08/11/2026 | 14 días | `feat: implementa cancelaciones y seguimiento de reembolsos` | Módulo 6: cancelación individual/de salida y constancia de devolución |
| 16 | 18/11/2026 | 10 días | `feat: completa la portada y navegacion publica de Andaria` | Módulo 7: home y experiencia integrada |
| 17 | 25/11/2026 | 7 días | `test: verifica el flujo turistico completo de Andaria` | Pruebas integrales, roles, compras, cupos y reembolsos |
| 18 | 07/12/2026 | 12 días | `docs: documenta el despliegue y la entrega de Andaria` | Verificación de publicación, HTTPS y guía final |

## Estado real al preparar el cronograma

- Base y módulo 1: archivos presentes y preparados en el índice, con algunas modificaciones posteriores.
- Módulo 2: archivos presentes, varios todavía sin seguimiento; su documentación describe el alcance entregado. Esta revisión de Git no volvió a ejecutar sus pruebas funcionales.
- Módulo 3: implementación y documentación presentes en el árbol de trabajo. El 21/09/2026 se verificaron las pruebas Go, las ocho pruebas de frontend y la comprobación de tipos; todavía debe prepararse como entrega Git independiente.
- Módulos 4–6: implementación y documentación presentes en el árbol de trabajo; todavía deben prepararse como entregas Git independientes.
- Módulo 7: portada pública, navegación y paneles por rol presentes en el árbol de trabajo y documentados; todavía debe prepararse como entrega Git independiente.
- Auditoría integral: pruebas de roles, agencias, atracciones, paquetes, compras, reembolsos, paneles y rutas públicas aprobadas el 25/09/2026. La compilación de contenedores quedó bloqueada por no estar disponible Docker Engine en este equipo.
- Despliegue: composición, variables y guía preparadas; la publicación real, HTTPS y Google OAuth quedan pendientes del dominio y del acceso al servidor.
- El flujo acordado no reserva mientras se transfiere: retiene al aceptar el comprobante. El cierre sin mínimo requiere confirmación de la agencia y los reembolsos tienen un plazo de 72 horas.

## Preparación de cada entrega

El índice actual contiene gran parte de la aplicación: un commit inmediato incluiría mucho más que la apertura del repositorio. Para usar entregas separadas habrá que seleccionar y revisar el contenido de cada una; algunos archivos actuales reúnen cambios de varios módulos y no basta con repartir carpetas. Cada entrega debe quedar coherente y verificable.

Antes de confirmar, revisar `git diff --cached --name-only` y `git diff --cached`. Incorporar únicamente archivos y cambios correspondientes a la entrega. No registrar funciones pendientes como terminadas ni ejecutar automáticamente los comandos del manual externo. Conservar las fechas reales de ejecución como evidencia de la migración desde el proyecto anterior.

En la preparación inicial no se hicieron commits ni pushes. Posteriormente se autorizó publicar las entregas existentes.

## Publicación de las entregas existentes

El 20/09/2026 se prepararon las entregas 2–10 a partir de archivos existentes. Se asignaron las fechas de referencia de la tabla a los commits por solicitud del propietario. Estas fechas no representan días de desarrollo o publicación verificados; se trata de la organización de una migración desde una base anterior.

Los títulos se ajustaron al contenido real: el backend del módulo 1 ya contiene Google, perfiles y administración; la entrega del 15/08 agrega sus pruebas y documentación. La API de agencias ya contiene configuración de pagos; la entrega del 02/09 incorpora la interfaz. No se ha simulado una implementación separada de funciones que ya estaban integradas.

Las entregas iniciales de infraestructura y esquema son parciales; el primer conjunto completo del módulo 1 queda reunido en la entrega 7. El módulo 2 queda reunido y documentado en la entrega 10. La implementación posterior de los módulos 3–7 permanece en el árbol de trabajo hasta preparar y revisar sus entregas; todavía no se ha publicado este conjunto.

Verificación de esta publicación: pruebas unitarias Go, go vet, cuatro pruebas frontend y comprobación de tipos aprobadas. Las pruebas de integración con PostgreSQL no se ejecutaron en esta sesión porque TEST_DATABASE_DSN no estaba configurada; las verificaciones anteriores descritas en los documentos de módulos son antecedentes, no resultados de esta sesión.
La compilación de producción del frontend también terminó correctamente en esta sesión.


## Actualización del cronograma: entregas del 14 al 24 de septiembre

Preparación y publicación de este lote: 26/09/2026. A solicitud del propietario, los siguientes commits usan fechas de referencia del 14 al 24/09/2026. No son evidencia de ejecución o publicación en esos días. Los informes conservan sus fechas reales de verificación, incluidas las del 25/09. Esta tabla reemplaza la planificación futura anterior de los módulos 3–7; las secciones anteriores se conservan como antecedentes.

Se agregan 21 commits, con un máximo de tres por día contando el commit del 14/09 que ya existía. Los esquemas se incorporan antes de las API. El backend integrado incluye funciones compartidas de los módulos 3–7; las entregas posteriores añaden interfaces, pruebas y documentación sin atribuir una segunda implementación a esas mismas funciones.

| Fecha de referencia | Mensaje |
| --- | --- |
| 2026-09-14 14:10 | `chore: excluye la documentacion local y habilita la plantilla de produccion` |
| 2026-09-14 17:25 | `fix: limita la configuracion comercial al encargado de agencia` |
| 2026-09-15 09:20 | `feat: incorpora el esquema de atracciones, horarios y fotografias` |
| 2026-09-15 13:40 | `feat: modela paquetes, programaciones y excepciones de salidas` |
| 2026-09-15 18:05 | `feat: incorpora compras, revision de pagos y reembolsos en la base de datos` |
| 2026-09-16 15:30 | `feat: integra las API de atracciones, paquetes, compras y reembolsos` |
| 2026-09-17 10:15 | `feat: configura mapas vectoriales y navegacion publica por rol` |
| 2026-09-17 16:40 | `feat: implementa el catalogo de atracciones, edicion y favoritos` |
| 2026-09-18 11:35 | `test: cierra la entrega de atracciones con pruebas de horarios y ubicacion` |
| 2026-09-18 17:10 | `feat: agrega la gestion de paquetes y su catalogo publico` |
| 2026-09-19 10:50 | `chore: agrega paquetes demostrativos para desarrollo` |
| 2026-09-19 16:20 | `docs: finaliza la entrega de paquetes, salidas y sus reglas` |
| 2026-09-20 14:30 | `feat: incorpora compra, comprobantes y seguimiento de reembolsos` |
| 2026-09-21 09:45 | `test: verifica los estados de compra y destinos de reembolso` |
| 2026-09-21 16:55 | `docs: cierra los modulos de compras, cancelaciones y reembolsos` |
| 2026-09-22 10:25 | `fix: conserva la sesion en SSR y corrige la proteccion de rutas` |
| 2026-09-22 17:35 | `fix: exige HTTPS y credenciales robustas en produccion` |
| 2026-09-23 11:20 | `feat: completa la portada y los paneles de Andaria por rol` |
| 2026-09-23 18:10 | `chore: prepara Docker y Nginx para el despliegue con HTTPS externo` |
| 2026-09-24 10:40 | `docs: incorpora las matrices y resultados de auditoria funcional` |
| 2026-09-24 17:20 | `docs: actualiza el alcance y registra el cierre de los modulos web` |

Estado del lote: módulos web 1–7 presentes. La finalización se refiere a su implementación y documentación; no acredita despliegue público ni ejecución real de pagos bancarios. Permanecen pendientes el servidor, dominio, HTTPS, OAuth real y las verificaciones finales de contenedores y restauración.

Verificación de esta sesión: pruebas Go y go vet aprobados; 16 pruebas frontend y comprobación de tipos aprobadas. Las pruebas de integración PostgreSQL se omiten sin TEST_DATABASE_DSN; los informes de auditoría describen ejecuciones anteriores y no se presentan como repetidas hoy. La carpeta `documentacion para el diplomando/`, los entornos reales y las copias locales siguen excluidos.


La compilación de producción del frontend también fue aprobada en esta sesión, con avisos de tamaño del mapa y de dependencias obsoletas que no impidieron generar la aplicación.


## Entrega de la app Android: 25/09 al 08/10/2026

Lote preparado el 08/10/2026 desde el árbol de trabajo existente. Las fechas asignadas organizan la incorporación del código y no acreditan desarrollo ni publicaciones en esos días. Se conserva el historial publicado hasta el 24/09 y se agregan 14 commits, uno por fecha. Esta sección actualiza el estado anterior que describía Flutter como pendiente.

La separación sigue las dependencias: estructura, núcleo móvil, contratos de API, componentes compartidos, pantallas y pruebas. Las primeras entregas móviles son parciales; el conjunto ejecutable queda integrado con catálogo, navegación y punto de entrada el 03/10. Las mejoras de aislamiento entre cuentas, reintentos y precios ya vienen integradas en sus archivos; no se inventan versiones defectuosas anteriores para presentarlas como correcciones separadas.

| Fecha asignada (Bolivia) | Commit |
| --- | --- |
| 2026-09-25 11:20 | `chore: incorpora la estructura Flutter y la configuracion Android` |
| 2026-09-26 16:10 | `feat: agrega el cliente API movil y las reglas de sesion y precios` |
| 2026-09-27 10:45 | `feat: habilita autenticacion Google para turistas desde Android` |
| 2026-09-28 15:35 | `fix: evita compras duplicadas y rechaza precios desactualizados` |
| 2026-09-29 09:50 | `feat: agrega galeria, comprobantes y contacto con la agencia` |
| 2026-09-30 17:15 | `feat: implementa acceso, registro y perfil del turista en la app` |
| 2026-10-01 11:30 | `feat: incorpora seguimiento de compras, cancelaciones y reembolsos` |
| 2026-10-02 16:40 | `feat: implementa la compra movil con revision de viajeros y comprobante` |
| 2026-10-03 10:25 | `feat: integra catalogo, favoritos y navegacion principal de Andaria` |
| 2026-10-04 15:10 | `test: verifica sesiones, cuentas, precios y experiencia movil` |
| 2026-10-05 12:40 | `test: agrega recorridos Android con datos aislados del backend` |
| 2026-10-06 17:05 | `ci: automatiza verificaciones de Android, API y web` |
| 2026-10-07 11:55 | `docs: conserva la auditoria tecnica previa a la aplicacion movil` |
| 2026-10-08 00:50 | `docs: documenta la app Android, su validacion y los limites de entrega` |

La auditoría técnica del 28/09 describe el corte web anterior y se conserva como documento histórico. El estado actual y las comprobaciones previas de emulador figuran en `docs/app-android.md` y `mobile/README.md`.

Pendientes externos: Google real con cliente y firma Android configurados, teclado y selector de fotos en teléfono físico, firma de distribución y publicación en tienda. Un APK debug verificado no constituye una publicación de producción. Las claves de firma, `local.properties`, entornos, APK, cachés y documentación privada del diplomado permanecen excluidos.



Verificación del lote el 08/10/2026: Flutter analyze sin incidencias; formato de 39 archivos sin cambios; 31 pruebas Flutter aprobadas; APK debug compilado correctamente; go test ./... -count=1 y go vet ./... aprobados; 16 pruebas web y comprobación de tipos aprobadas. No se repitieron los recorridos de emulador ni las integraciones con PostgreSQL, pues TEST_DATABASE_DSN no estaba configurada. Las versiones de acciones y Flutter del workflow existen en sus repositorios oficiales; su ejecución remota queda pendiente de comprobar después del push.
