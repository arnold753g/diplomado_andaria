# Lectura de comprobantes en Flutter web

Revisión del 28/10/2026. El fallo reportado ocurre al ejecutar la app Flutter en Chrome con emulación móvil, antes de enviar la compra al servidor.

## Causa y corrección

`ProofImage.validate` consultaba `ImageDescriptor.width` y `height` sobre una imagen codificada. Flutter web no implementa esas propiedades y lanza `UnsupportedError`; el manejo del error lo convertía en «No se pudo leer la imagen. Selecciona otro PNG o JPG.», incluso para un PNG válido.

Se utiliza `instantiateImageCodecWithSize` y su callback de dimensiones originales, compatible con Flutter web y Android. La API administra el buffer; el lector libera el codec y la imagen decodificada. Se conservan los bytes originales, las firmas PNG/JPEG, el límite de 5 MB, el máximo de 4096 × 4096 píxeles y el rechazo de imágenes dañadas. La corrección se aplica al lector compartido de comprobantes y otras imágenes de pago.

## Verificación

- Reproducción antes del cambio: un PNG sintético válido muestra exactamente el mensaje reportado en el navegador.
- Después del cambio, navegador Chromium integrado, pantalla de 390 × 844: selección real mediante el selector de archivos y vista previa correctas para PNG y JPG sintéticos. Una imagen de 4097 píxeles de ancho, otra de 4097 de alto y un PNG dañado se rechazan con sus mensajes correspondientes.
- `flutter test --no-pub --reporter expanded`: 45 pruebas aprobadas, incluidas conservación de bytes/MIME, ancho permitido de 4096, exceso de ancho/alto, archivos dañados, formato y peso.
- `flutter analyze --no-pub`: sin incidencias.
- Compilación web de la pantalla aislada: correcta. Las pruebas usan imágenes sintéticas sin datos personales y no envían compras.

El ejecutor `flutter test --platform chrome` no completó el arranque en este entorno: el snapshot del SDK buscó recursos de `test` en una ruta de la máquina de compilación y la ejecución desde fuentes quedó detenida al iniciar la suite. La comprobación del navegador se realizó con una compilación web aislada del componente real; no se declara aprobada esa suite de Chrome.

Evidencias: [antes](evidencias/comprobante-web-2026-10-28/antes.png), [JPG seleccionado](evidencias/comprobante-web-2026-10-28/despues.jpg.png) y [PNG en pantalla móvil](evidencias/comprobante-web-2026-10-28/despues-movil.png).

## Aplicar a la sesión de desarrollo

En la terminal donde está ejecutándose `flutter run -d chrome`, realizar **Hot Restart** con `R` y volver a seleccionar el comprobante. Si la ejecución ya terminó, iniciarla otra vez con la configuración habitual. Una recarga del navegador por sí sola puede seguir usando la compilación anterior.

No se requieren cambios en la base de datos, Google Cloud ni el VPS para esta corrección. No se ha generado una nueva APK ni publicado una versión en producción en esta revisión.
