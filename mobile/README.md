# Andaria Android

App Flutter + Dart para turistas. Comparte la API Go y los datos de la web. El catálogo es público; favoritos, perfil y compras requieren una cuenta con rol `turista`.

## Ejecutar

La entrega se verificó con Flutter 3.44.8 y Dart 3.12.2. Necesitas el SDK Flutter disponible en `PATH`, Android SDK y el JDK usado por Android Studio. Comprueba tu instalación con `flutter doctor -v`.

Desde `mobile/`:

```powershell
flutter pub get
flutter run -d emulator-5554
```

El emulador usa `http://10.0.2.2:8080/api/v1` para acceder al backend del equipo. Inicia el backend desde `backend/` con `go run ./cmd/api`. Aplica antes las migraciones con `go run ./cmd/migrate` si no está habilitada la migración automática.

### Vista de desarrollo en Chrome

Desde `mobile/`, con el backend en el puerto 8080 y la web en el 3000:

```powershell
flutter run -d chrome --web-hostname=localhost --web-port=7357 --dart-define=API_BASE_URL=http://localhost:8080/api/v1
```

En el `.env` de la raíz, conserva el origen de la web y agrega el de Flutter:

```dotenv
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:7357
```

Reinicia el backend después de cambiar esta variable. Si Flutter ya estaba ejecutándose, detén esa ejecución y vuelve a iniciarla con el comando anterior; cambiar `--dart-define` requiere reiniciar. Abre `http://localhost:7357`, manteniendo `localhost` tanto en la app como en la API. `10.0.2.2` corresponde al emulador Android y no sirve para Chrome en el equipo.

La URL predeterminada ahora se adapta al destino: Chrome usa `localhost`, Android conserva `10.0.2.2`. El puerto fijo evita que CORS rechace un origen aleatorio. Chrome conserva la cookie HttpOnly y la envía con credenciales; Dart solo conserva el token CSRF y la cuenta en memoria. En Chrome la sesión se comparte con la web cuando ambas utilizan el mismo backend y navegador; para usar otra cuenta, cierra la sesión o utiliza otro perfil del navegador. El acceso Google nativo se muestra únicamente en Android; la vista Chrome sirve para probar el catálogo y los recorridos con correo y contraseña.

Pruebas específicas de sesión del navegador:

```powershell
flutter test --platform chrome test/browser_session_test.dart
```

En la instalación local Flutter 3.44.8, el ejecutor de estas tres pruebas se bloqueó antes de iniciarlas por un `Null check operator used on a null value` en `host.dart.js` del host de pruebas (elemento `#play` ausente). Quedan pendientes en un ejecutor compatible. Se verificaron por separado la compilación web en profile, el catálogo conectado en un navegador y las 31 pruebas existentes; `flutter analyze` pasó sin errores.

Para un teléfono físico, usa una dirección accesible en la red local:

```powershell
flutter run --dart-define=API_BASE_URL=http://192.168.1.20:8080/api/v1
```

El HTTP de desarrollo está permitido solo en la variante debug. La app exige HTTPS cuando se compila en release. No uses credenciales de producción contra un backend HTTP.

## Verificar y compilar

```powershell
flutter analyze
flutter test
flutter build apk --debug
```

APK de pruebas: `build/app/outputs/flutter-apk/app-debug.apk`.

Para la prueba de integración con el catálogo real, inicia el backend y un emulador. La prueba solo consulta datos públicos y abre el formulario de acceso:

```powershell
flutter drive --driver=test_driver/integration_test.dart --target=integration_test/guest_flow_test.dart -d emulator-5554
```

Las capturas de esta prueba se guardan en `../.local/mobile-preview/` y no se versionan.

Para una entrega release, copia `android/key.properties.example` a `android/key.properties` y configura un certificado propio. El archivo y los certificados quedan excluidos de Git. La compilación release exige esta configuración y una `API_BASE_URL` HTTPS; no utiliza automáticamente la firma debug.

## Prueba completa con datos aislados

El recorrido de compras usa un servidor de pruebas en el puerto 8081. El helper Go crea un esquema PostgreSQL temporal con turistas, agencia, lugar y salida; lo elimina al terminar. Usa una base local de pruebas para `TEST_DATABASE_DSN`.

Desde `backend/`, en una terminal:

```powershell
$env:TEST_DATABASE_DSN = 'host=localhost port=5432 user=postgres password=TU_CLAVE_LOCAL dbname=TU_BASE_DE_PRUEBAS sslmode=disable'
$env:ANDARIA_ANDROID_FIXTURE = '1'
$env:ANDARIA_ANDROID_STOP_FILE = [IO.Path]::GetFullPath((Join-Path (Get-Location) '../.local/android-fixture-stop'))
go test ./internal/httpapi -run '^TestAndroidFixture$' -v -count=1 -timeout=30m
```

Elige un nombre de archivo de parada que todavía no exista. Cuando aparezca `Android fixture ready`, desde `mobile/` en otra terminal:

```powershell
flutter drive --driver=test_driver/integration_test.dart --target=integration_test/purchase_flow_test.dart -d emulator-5554 --dart-define=API_BASE_URL=http://10.0.2.2:8081/api/v1
```

Al finalizar, desde la raíz del repositorio crea el archivo de parada elegido:

```powershell
New-Item -ItemType File -Path .local/android-fixture-stop
```

Espera a que el proceso Go termine y limpie el esquema. El helper también se detiene automáticamente a los 25 minutos. No detengas el proceso a la fuerza durante esta prueba. Las imágenes de comprobante se simulan; se verifica su envío, procesamiento y consulta a través de la API real, sin realizar transferencias bancarias.

Con el mismo servidor aislado todavía en ejecución, puedes validar registro, cambio de contraseña y revocación de sesiones:

```powershell
flutter drive --driver=test_driver/integration_test.dart --target=integration_test/account_flow_test.dart -d emulator-5554 --dart-define=API_BASE_URL=http://10.0.2.2:8081/api/v1
```

Esta prueba crea una cuenta temporal, abre una segunda sesión y comprueba que ambas quedan revocadas al cambiar la contraseña. Simula la entrada de texto para evitar interferencias del teclado del emulador; conserva los formularios, validaciones y solicitudes HTTP reales. Ejecuta ambas pruebas antes de crear el archivo de parada. El uso del teclado y del selector de imágenes del sistema requiere también comprobación manual en un teléfono físico.

## Funciones

- Explorar experiencias y lugares, buscar, filtrar por departamento y ordenar paquetes.
- Consultar fotografías con ampliación, itinerarios con acceso a los lugares, precios, horarios, ubicación y salidas disponibles.
- Acceder por correo/contraseña o Google; registrarse cuando el backend lo permite.
- Guardar atracciones favoritas y editar datos personales/contraseña.
- Comprar una salida con viajeros nacionales, extranjeros y menores; revisar el comprobante ampliado antes de enviarlo; pagar por QR o transferencia y adjuntar PNG/JPG de hasta 5 MB y 4096 × 4096 píxeles.
- Consultar compras y comprobantes, corregir respaldos, cancelar según la política y proporcionar destino de devolución.
- Consultar el plazo y comprobante de reembolso.
- Consultar el contacto comercial de la agencia, abrir el teléfono o redactar un correo con la referencia de compra.
- Revisar los datos del comprador antes del envío y volver a consultar la tarifa o las salidas si el servidor detecta un cambio de precio o disponibilidad.

Español, bolivianos y fechas mostradas en UTC−4, independientemente de la zona horaria del teléfono. Tema claro con Outfit, superficies neutras y acentos verde lima de Andaria. Outfit se incluye localmente con su licencia OFL.

## Google en Android

El backend publica su identificador OAuth web en `GET /api/v1/auth/mobile/options`; el secreto OAuth no se entrega a la app. Android obtiene un ID token y la API lo valida con la librería oficial de Google antes de emitir una sesión Andaria.

Para habilitar el acceso real:

1. Mantén la configuración Google completa del backend: `GOOGLE_CLIENT_ID` (cliente OAuth web), `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL` y `FRONTEND_URL` coherente con `ALLOWED_ORIGINS`. Android recibe únicamente el identificador público.
2. Registra un cliente OAuth Android en el mismo proyecto, con el paquete `bo.andaria.andaria_mobile` y la huella SHA-1 del certificado que firma el APK.
3. Para desarrollo, consulta la huella mediante `./gradlew.bat signingReport`, ejecutado desde `mobile/android/` con un JDK configurado en `JAVA_HOME`.
4. Registra también la firma de distribución cuando prepares esa entrega.

No hace falta incorporar `google-services.json` para este flujo: se pasa el identificador web como `serverClientId`. La vinculación de Google a una cuenta ya existente conserva el flujo explícito de la web.

## Organización

- `lib/core/`: API, sesión, formatos, cálculo de grupos, tema y acceso Google.
- `lib/features/`: catálogo, autenticación, perfil, compra y seguimiento.
- `lib/shared/`: componentes, selección de imágenes y protección de pantallas por cuenta.
- `test/`: pruebas de sesión, precios, navegación y aislamiento entre cuentas.
- `integration_test/`: recorrido Android contra el catálogo real.

Las cookies se guardan en almacenamiento seguro Android; el token CSRF y los datos de sesión se recuperan del servidor. Una respuesta de una sesión anterior no puede reemplazar la cuenta actual. Las pantallas privadas dejan de mostrar sus datos si cambia la cuenta.

Las solicitudes de compra usan `Idempotency-Key`. Una clave pendiente se conserva cifrada y se reutiliza si se vuelve a enviar el mismo grupo, salida, precio y comprobante. Las llamadas simultáneas comparten la obtención de esa clave; cambiar de cuenta durante la confirmación detiene el envío. Antes de intentar otra compra tras un error de conexión, consulta **Mis compras**. El cupo se retiene al aceptar el comprobante, no durante la transferencia; pago enviado y pago confirmado son estados distintos.

## Límites de esta entrega

El acceso real con Google depende de registrar paquete y firma en Google Cloud. La recuperación de contraseña, las notificaciones push, el uso sin conexión y la eliminación de cuenta no están implementados. La publicación en tiendas requiere preparar firma, privacidad y requisitos de distribución. El APK debug es para pruebas.

## Integración continua

`.github/workflows/quality.yml` prepara tres verificaciones en GitHub Actions: análisis, pruebas y APK Android; backend con PostgreSQL temporal; pruebas y tipos de la web. El workflow se ejecutará al subir los cambios o abrir una solicitud de revisión. Su ejecución remota todavía debe comprobarse en GitHub.
