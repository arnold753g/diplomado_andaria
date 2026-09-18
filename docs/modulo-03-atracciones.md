# Módulo 3: atracciones turísticas y favoritos

## Alcance

- El administrador crea atracciones, asigna un encargado por atracción y puede reasignarlas o desactivarlas.
- Un encargado puede gestionar varias atracciones. Puede editar y publicar únicamente las que tiene asignadas; no puede crearlas, cambiar encargados ni modificar el estado administrativo.
- Ficha: nombre, descripción, categoría, departamento, municipio, dirección o referencia, coordenadas opcionales, horarios, entrada informativa en bolivianos, recomendaciones y teléfono.
- Galería de hasta seis fotografías PNG/JPEG de 5 MB y 4096 × 4096 píxeles cada una. La primera funciona como portada. Se pueden quitar o elegir otra portada al editar.
- El servidor valida el contenido real de las imágenes y las vuelve a codificar en JPEG, descartando metadatos. Las fotos se guardan en PostgreSQL junto con su atracción.
- Borradores: para publicar se exige una descripción de al menos 20 caracteres y una fotografía.
- Catálogo público con búsqueda por nombre/localidad, filtros por departamento y categoría y paginación. Solo muestra atracciones activas y publicadas.
- Ficha pública con galería, información para la visita y mapa interactivo con marcador si se completaron las coordenadas.
- Favoritos exclusivos para turistas: guardar, consultar y quitar. Son persistentes e independientes por usuario.
- Una atracción oculta o inactiva desaparece del catálogo y de la lista visible de favoritos; el vínculo se conserva y reaparece si vuelve a publicarse. Sus fotografías tampoco pueden consultarse públicamente.
- Protección de versiones para evitar sobrescribir cambios concurrentes. Reasignar una atracción retira inmediatamente el acceso del encargado anterior.
- Se bloquea el cambio de rol de un encargado mientras tenga atracciones asignadas.

No incluye aún paquetes, entradas vendidas, reseñas ni una administración separada de categorías. La clasificación usa las cuatro categorías y 33 subcategorías del catálogo de Andaria. El modelo admite los nueve departamentos de Bolivia; no se cargan atractivos reales automáticamente.

## Recorrido de revisión

1. Ingresar como administrador y crear un usuario con rol **Encargado de atracción** desde Usuarios.
2. Abrir **Atracciones → Crear atracción**, completar los datos iniciales y asignar ese encargado. Puede guardarse como borrador sin foto.
3. Ingresar con el encargado y abrir **Mis atracciones**. Completar la ficha, agregar fotografías y marcar **Publicar en el catálogo**; guardar.
4. Consultar `http://127.0.0.1:3000/attractions` sin necesidad de iniciar sesión.
5. Ingresar como turista, abrir la ficha y guardar el lugar. Revisar **Mis favoritos** y quitarlo si se desea.
6. Despublicar desde el encargado o desactivar desde el administrador y comprobar que ya no aparece públicamente.

## API compartida con la futura app móvil

- Público: `GET /api/v1/attractions`, `/options`, `/{id}` y `/{id}/photos/{photo}`.
- Administrador: `/api/v1/admin/attractions` y sus rutas de detalle, fotografías y selección de encargados.
- Encargado: `GET /api/v1/managed-attractions`, `GET/PUT /{id}` y fotografías privadas.
- Turista: `GET /api/v1/me/favorites`; `GET/PUT/DELETE /{id}` para consultar, guardar y quitar un favorito. PUT y DELETE son idempotentes.

## Verificación

- Migración `000004_attractions.sql` aplicada; las pruebas de integración usan esquemas temporales aislados.
- Pruebas de autenticación, CSRF, roles, acceso entre encargados, publicación, visibilidad de fotos, datos públicos sin información del encargado, filtros y favoritos.
- Pruebas de imágenes inválidas, coordenadas, precios y apropiación indebida de fotos; una edición inválida revierte todos sus cambios.
- Pruebas concurrentes: favoritos repetidos no se duplican y dos ediciones con la misma versión no sobrescriben cambios.
- Suite Go, análisis estático, comprobación de tipos de frontend, pruebas de frontend y compilación de producción.
- CORS permite PUT y DELETE desde los orígenes configurados, con prueba de preflight que sigue rechazando orígenes ajenos.
- Revisión en navegador: creación de borrador por administrador, acceso con encargado, carga de fotografía, publicación, catálogo, ficha pública y guardado/eliminación de favoritos con turista. Los datos temporales de estas pruebas se retiraron al finalizar.
- Nginx permite hasta 42 MB únicamente en las rutas de gestión de atracciones para las galerías codificadas; las demás rutas conservan su límite previo.

Se mantiene el estilo visual de la base. La ejecución local usa frontend en `127.0.0.1:3000` y backend en `127.0.0.1:8082`.

## Actualización: clasificación, horarios, temporada y mapa

- Migración `000005_attraction_classification_schedule.sql`: catálogo copiado de `V2 proy/backend/internal/seeds/categorias_seed.go`: Enoturismo (5), Cultural (8), Natural (14) y Deportivo (6).
- Se seleccionan entre una y cuatro subcategorías, incluso de categorías distintas. La primera es la principal; el usuario puede cambiarla. Los filtros consideran también las subcategorías complementarias.
- Se mantiene la nomenclatura original, incluido «Posas». Los identificadores son propios de esta base; no dependen de los IDs de la base de Andaria.
- Horarios seleccionables: por confirmar, horario de apertura/cierre o 24 horas; días de lunes a domingo. Un mismo intervalo se aplica a los días elegidos, igual que en Andaria. Las horas se almacenan como HH:mm local, sin conversiones de zona horaria. Un cierre menor que la apertura indica el día siguiente; horas iguales se rechazan y se ofrece «24 horas».
- Temporada recomendada: por confirmar, todo el año o meses de inicio/fin inclusivos. Admite temporadas que cruzan el año y un solo mes. Es la mejor época de visita, no una regla automática de cierre ni de publicación.
- Migración conservadora: no se inventan subcategorías ni horarios para registros previos. La categoría y el horario de texto anterior se conservan y se muestran para que el encargado los estructure al editar.
- Mapa: Leaflet 1.9.4 + MapLibre GL Leaflet + datos vectoriales de OpenStreetMap servidos por OpenFreeMap. El estilo local aplica la paleta Andaria nocturno a agua, vegetación, edificios, vías, límites y rótulos. Centro inicial de referencia en Tarija, sin guardar coordenadas hasta que el usuario elija una ubicación. Clic, marcador arrastrable, selección del centro del mapa y edición manual opcional; respeta los permisos de solo lectura.
- Mientras carga el estilo vectorial se muestra la capa raster de OpenStreetMap. Si WebGL o el proveedor vectorial fallan, esa capa permanece disponible y la selección del marcador continúa funcionando.
- Estilo configurable con `NUXT_PUBLIC_MAP_STYLE_URL`; respaldo configurable con `NUXT_PUBLIC_MAP_TILE_URL`. Si se cambia el proveedor, también deben actualizarse `connect-src` e `img-src` en Nginx. La atribución permanece visible; no se precargan mapas ni se ofrecen descargas offline.
- El mapa integrado no calcula rutas ni incluye imágenes satelitales. La ficha pública ofrece **Cómo llegar con Google Maps** mediante una URL universal con las coordenadas exactas; en móvil abre la aplicación si está disponible y conserva la versión web como respaldo. No requiere una clave de API.

### Opciones de mapas

| Alternativa | Tecnología | Consideración |
| --- | --- | --- |
| Calles estilizadas (implementada) | Leaflet + MapLibre + OpenFreeMap/OSM | Conserva el marcador de Leaflet y permite una identidad visual propia mediante teselas vectoriales sin API key. |
| Calles/satélite/híbrido/terreno | Google Maps JavaScript API | Requiere clave y facturación habilitada. |
| Vectorial y estilos avanzados | MapLibre GL JS + proveedor de mapas | Biblioteca abierta; el alojamiento y las capas cartográficas se contratan o gestionan aparte. |

Fuentes: [Leaflet](https://leafletjs.com/), [política de tiles de OpenStreetMap](https://operations.osmfoundation.org/policies/tiles/), [tipos de mapa Google](https://developers.google.com/maps/documentation/javascript/maptypes), [facturación de Google Maps](https://developers.google.com/maps/documentation/javascript/usage-and-billing), [MapLibre](https://maplibre.org/projects/gl-js/).
