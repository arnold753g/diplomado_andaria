# Módulo 7 — home pública y paneles por rol

Estado: implementado en el árbol de trabajo y verificado el 24 de septiembre de 2026.

## Objetivo

Cerrar la experiencia web con una portada pública que permita descubrir el contenido sin iniciar sesión y paneles privados que presenten a cada rol sus próximas tareas. Este módulo reutiliza los datos reales de atracciones, paquetes, compras y reembolsos; no introduce un segundo catálogo ni contenido destacado administrado manualmente.

## Portada pública

La ruta `/` presenta:

- el mensaje «Explora Bolivia, comenzando por Tarija»;
- el paquete con la salida comprable más próxima;
- hasta tres paquetes con venta abierta, ordenados por próxima salida;
- hasta tres atracciones publicadas recientes;
- una explicación breve del flujo de exploración, compra y seguimiento;
- accesos al catálogo, registro e inicio de sesión.

Si no existe contenido publicado, cada sección muestra un estado vacío y conserva un acceso útil al catálogo. Las imágenes proceden de las fotografías cargadas en Andaria. La interfaz mantiene el tema central, navegación adaptable, jerarquía de encabezados, textos alternativos, objetivos táctiles y soporte para reducción de movimiento.

## Paneles privados

`GET /api/v1/me/dashboard` exige sesión y responde de acuerdo con el rol:

| Rol | Información principal |
| --- | --- |
| Turista | favoritos, compras activas, reembolsos pendientes/completados y próxima compra |
| Encargado de agencia | paquetes, pagos por revisar, decisiones por cupo mínimo, reembolsos pendientes, próximos a vencer y atrasados |
| Encargado de atracción | atracciones asignadas, publicadas, borradores e inactivas |

El administrador usa `GET /api/v1/admin/dashboard`, ampliado con usuarios activos, agencias publicadas, atracciones publicadas, paquetes publicados y compras totales. Un administrador no puede consultar el panel personal de otro rol.

Los conteos de agencia se limitan a la agencia asignada y los de atracciones a las asignaciones del encargado. El backend es la fuente de verdad para este alcance.

## Rutas de interfaz

- `/`: portada pública.
- `/app`: panel de turista, encargado de agencia o encargado de atracción según la sesión.
- `/admin`: panel global del administrador.
- `/attractions` y `/packages`: catálogos públicos.

## Verificación

- Prueba de integración de acceso anónimo, denegación al rol incorrecto y conteos aislados para los cuatro roles.
- Pruebas de navegación pública y por rol.
- Pruebas completas y análisis estático de Go.
- Pruebas, comprobación de tipos y compilación de producción de Nuxt.
- Revisión visual de la portada en anchos móvil y escritorio.

## Límites de esta entrega

- La selección de contenido es automática; no existe orden editorial manual.
- No se calculan recomendaciones personalizadas ni distancia geográfica al visitante.
- Los contadores se actualizan al volver a cargar o entrar al panel; no se usa actualización en tiempo real.
