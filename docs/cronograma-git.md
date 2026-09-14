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
- Módulos 3–7: pendientes según el alcance y los archivos revisados. Las fechas son objetivos propuestos, no entregas ya realizadas.
- La retención de cupos de diez minutos, el plazo de revisión de comprobantes y la fecha límite del mínimo de salida siguen pendientes de decisión.

## Preparación de cada entrega

El índice actual contiene gran parte de la aplicación: un commit inmediato incluiría mucho más que la apertura del repositorio. Para usar entregas separadas habrá que seleccionar y revisar el contenido de cada una; algunos archivos actuales reúnen cambios de varios módulos y no basta con repartir carpetas. Cada entrega debe quedar coherente y verificable.

Antes de confirmar, revisar `git diff --cached --name-only` y `git diff --cached`. Incorporar únicamente archivos y cambios correspondientes a la entrega. No registrar funciones pendientes como terminadas ni ejecutar automáticamente los comandos del manual externo. Conservar las fechas reales de ejecución como evidencia de la migración desde el proyecto anterior.

En la preparación inicial no se hicieron commits ni pushes. Posteriormente se autorizó publicar las entregas existentes.

## Publicación de las entregas existentes

El 20/09/2026 se prepararon las entregas 2–10 a partir de archivos existentes. Se asignaron las fechas de referencia de la tabla a los commits por solicitud del propietario. Estas fechas no representan días de desarrollo o publicación verificados; se trata de la organización de una migración desde una base anterior.

Los títulos se ajustaron al contenido real: el backend del módulo 1 ya contiene Google, perfiles y administración; la entrega del 15/08 agrega sus pruebas y documentación. La API de agencias ya contiene configuración de pagos; la entrega del 02/09 incorpora la interfaz. No se ha simulado una implementación separada de funciones que ya estaban integradas.

Las entregas iniciales de infraestructura y esquema son parciales; el primer conjunto completo del módulo 1 queda reunido en la entrega 7. El módulo 2 queda reunido y documentado en la entrega 10. No se publican implementaciones de los módulos 3–7.

Verificación de esta publicación: pruebas unitarias Go, go vet, cuatro pruebas frontend y comprobación de tipos aprobadas. Las pruebas de integración con PostgreSQL no se ejecutaron en esta sesión porque TEST_DATABASE_DSN no estaba configurada; las verificaciones anteriores descritas en los documentos de módulos son antecedentes, no resultados de esta sesión.
La compilación de producción del frontend también terminó correctamente en esta sesión.
