import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'models.dart';
import 'http_platform.dart';

abstract class SessionStorage {
  Future<String?> read();
  Future<void> write(String value);
  Future<void> clear();
}

// Web credentials stay in the browser's HttpOnly cookie jar.
class BrowserSessionStorage implements SessionStorage {
  @override
  Future<String?> read() async => null;
  @override
  Future<void> write(String value) async {}
  @override
  Future<void> clear() async {}
}

class SecureSessionStorage implements SessionStorage {
  SecureSessionStorage(String baseUrl) : key = 'andaria.session.$baseUrl';
  final String key;
  final FlutterSecureStorage storage = const FlutterSecureStorage();
  @override
  Future<String?> read() => storage.read(key: key);
  @override
  Future<void> write(String value) => storage.write(key: key, value: value);
  @override
  Future<void> clear() => storage.delete(key: key);
}

class ApiFailure implements Exception {
  const ApiFailure(
    this.message, {
    this.code = '',
    this.fields = const {},
    this.status,
  });
  final String message, code;
  final Map<String, String> fields;
  final int? status;
  @override
  String toString() => message;
}

const errorTranslations = {
  'INVALID_CREDENTIALS': 'El correo o la contraseña no son correctos.',
  'ACCOUNT_INACTIVE': 'Esta cuenta está desactivada.',
  'ACCOUNT_LOCKED': 'Demasiados intentos. Espera antes de volver a ingresar.',
  'EMAIL_EXISTS': 'Ya existe una cuenta con ese correo.',
  'REGISTRATION_DISABLED': 'El registro no está disponible en este momento.',
  'PASSWORD_POLICY': 'La contraseña no cumple los requisitos de seguridad.',
  'PASSWORD_REUSED': 'Elige una contraseña diferente de la actual.',
  'CURRENT_PASSWORD_INVALID': 'La contraseña actual no es correcta.',
  'VALIDATION_ERROR': 'Revisa los datos del formulario.',
  'UNAUTHORIZED': 'Inicia sesión para continuar.',
  'INVALID_SESSION': 'Tu sesión venció. Inicia sesión nuevamente.',
  'SESSION_EXPIRED': 'Tu sesión venció. Inicia sesión nuevamente.',
  'FORBIDDEN': 'Tu cuenta no tiene permiso para esta operación.',
  'CSRF_INVALID': 'No se pudo validar la sesión. Vuelve a iniciar sesión.',
  'INTERNAL_ERROR': 'No se pudo completar la operación. Inténtalo nuevamente.',
};

class AndariaApi extends ChangeNotifier {
  AndariaApi({String? baseUrl, Dio? dio, SessionStorage? storage}) {
    final url =
        (baseUrl ??
                const String.fromEnvironment(
                  'API_BASE_URL',
                  defaultValue: kIsWeb
                      ? 'http://localhost:8080/api/v1'
                      : 'http://10.0.2.2:8080/api/v1',
                ))
            .replaceFirst(RegExp(r'/$'), '');
    final uri = Uri.parse(url);
    if (!uri.hasAuthority ||
        !['http', 'https'].contains(uri.scheme) ||
        (kReleaseMode && uri.scheme != 'https')) {
      throw ArgumentError(
        'Configura API_BASE_URL con una URL válida; producción requiere HTTPS.',
      );
    }
    client = dio ?? Dio();
    if (dio == null) configurePlatformClient(client);
    client.options = BaseOptions(
      baseUrl: url,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 30),
      sendTimeout: const Duration(seconds: 45),
      followRedirects: false,
    );
    sessionStorage =
        storage ??
        (kIsWeb ? BrowserSessionStorage() : SecureSessionStorage(url));
  }
  late final Dio client;
  late final SessionStorage sessionStorage;
  String? _cookie, _csrf;
  int _sessionEpoch = 0;
  Future<void>? _restoring;
  Timer? _expiration;
  final Stopwatch _sessionClock = Stopwatch();
  Duration? _absoluteRemaining;
  Duration _idleTimeout = const Duration(minutes: 30);
  DateTime? _serverTimestamp;
  Json? user;
  bool get signedIn =>
      user != null && (kIsWeb || _cookie != null) && _csrf != null;
  String get fullName => user == null
      ? ''
      : [
          user!['first_name'],
          user!['last_name'],
        ].whereType<String>().where((value) => value.isNotEmpty).join(' ');
  String? sessionNotice;

  Future<void> restore() => _restoring ??= _restoreSession().whenComplete(() {
    _restoring = null;
  });

  Future<void> _restoreSession() async {
    final epoch = _sessionEpoch;
    try {
      final stored = await sessionStorage.read();
      if (epoch != _sessionEpoch) return;
      _cookie = stored;
      if (!kIsWeb && _cookie == null) return;
      final data = await request('/auth/session');
      if (epoch != _sessionEpoch) return;
      await _acceptSession(Map<String, dynamic>.from(data as Map));
    } on ApiFailure catch (e) {
      if (epoch != _sessionEpoch) return;
      if (e.status == 401 || e.status == 403) {
        sessionNotice = kIsWeb && user == null && e.status == 401
            ? null
            : e.status == 401
            ? 'Tu sesión venció. Inicia sesión nuevamente.'
            : e.message;
        try {
          await clear();
        } catch (_) {
          /* Local private state was already cleared. */
        }
      } else {
        sessionNotice =
            'No pudimos recuperar tu sesión. Comprueba tu conexión.';
        notifyListeners();
      }
    } catch (_) {
      if (epoch == _sessionEpoch) {
        sessionNotice = 'No pudimos recuperar tu sesión.';
        notifyListeners();
      }
    }
  }

  Future<void> _acceptSession(Json data) async {
    _csrf = data['csrf_token'] as String?;
    final account = Map<String, dynamic>.from(data['user'] as Map);
    if (account['role'] != 'turista') {
      try {
        await request('/auth/logout', method: 'POST');
      } catch (_) {}
      sessionNotice =
          'Esta app es para turistas. Usa la web para gestionar otros roles.';
      await clear();
      throw const ApiFailure(
        'Esta app es para turistas. Usa la web para gestionar tu agencia o administración.',
        code: 'TOURIST_REQUIRED',
      );
    }
    user = account;
    sessionNotice = null;
    final expires = DateTime.tryParse(
      data['session_expires_at']?.toString() ?? '',
    );
    _absoluteRemaining = expires?.toUtc().difference(
      _serverTimestamp ?? DateTime.now().toUtc(),
    );
    final idleSeconds = integer(data['idle_timeout_seconds']);
    _idleTimeout = Duration(seconds: idleSeconds > 0 ? idleSeconds : 1800);
    _sessionClock.reset();
    _sessionClock.start();
    _scheduleExpiration();
    notifyListeners();
  }

  void _scheduleExpiration() {
    _expiration?.cancel();
    if (!signedIn || _absoluteRemaining == null) return;
    final remaining = _absoluteRemaining! - _sessionClock.elapsed;
    final delay = remaining < _idleTimeout ? remaining : _idleTimeout;
    _expiration = Timer(delay.isNegative ? Duration.zero : delay, () {
      sessionNotice = 'Tu sesión venció. Inicia sesión nuevamente.';
      unawaited(clear().catchError((Object _) {}));
    });
  }

  @override
  void dispose() {
    _expiration?.cancel();
    _sessionClock.stop();
    client.close();
    super.dispose();
  }

  Future<void> login(String email, String password) async {
    await clear();
    final data = await request(
      '/auth/login',
      method: 'POST',
      body: {'email': email.trim(), 'password': password},
    );
    await _acceptSession(Map<String, dynamic>.from(data as Map));
  }

  Future<void> register(Json payload) async {
    await request('/auth/register', method: 'POST', body: payload);
  }

  Future<void> loginGoogle(String idToken) async {
    await clear();
    final data = await request(
      '/auth/google/mobile',
      method: 'POST',
      body: {'id_token': idToken},
    );
    await _acceptSession(Map<String, dynamic>.from(data as Map));
  }

  Future<void> logout() async {
    try {
      if (kIsWeb || _cookie != null) {
        await request('/auth/logout', method: 'POST');
      }
    } finally {
      await clear();
    }
  }

  Future<void> clear() async {
    _expiration?.cancel();
    _sessionClock.stop();
    _absoluteRemaining = null;
    _sessionEpoch++;
    _cookie = _csrf = null;
    user = null;
    notifyListeners();
    await sessionStorage.clear();
  }

  Future<dynamic> request(
    String path, {
    String method = 'GET',
    Json? body,
    Json? query,
    bool authenticated = false,
    String? idempotencyKey,
  }) async {
    if (authenticated && !signedIn) {
      throw const ApiFailure('Inicia sesión para continuar.', status: 401);
    }
    final epoch = _sessionEpoch;
    try {
      final response = await client.request<dynamic>(
        path,
        data: body,
        queryParameters: query,
        options: Options(
          method: method,
          headers: {
            if (!kIsWeb && _cookie != null) 'Cookie': _cookie,
            if (_csrf != null && method != 'GET') 'X-CSRF-Token': _csrf,
            'Idempotency-Key': ?idempotencyKey,
          },
        ),
      );
      if (epoch != _sessionEpoch &&
          (authenticated || path.startsWith('/auth/'))) {
        throw const ApiFailure(
          'La sesión cambió. Vuelve a intentar la operación.',
          code: 'SESSION_CHANGED',
        );
      }
      for (final cookie
          in epoch == _sessionEpoch
              ? sessionCookies(response.headers)
              : <({String value, bool revoked})>[]) {
        if (cookie.revoked) {
          try {
            await clear();
          } catch (_) {
            // The server revoked this cookie; private memory is already clear.
            // A storage cleanup error must not misreport a confirmed mutation.
          }
        } else {
          _cookie = cookie.value;
          await sessionStorage.write(_cookie!);
        }
      }
      if (response.data is! Map) {
        throw const ApiFailure('La respuesta del servidor no es válida.');
      }
      final envelope = Map<String, dynamic>.from(response.data as Map);
      if (envelope['success'] != true) {
        throw const ApiFailure('La respuesta del servidor no es válida.');
      }
      _serverTimestamp = DateTime.tryParse(
        envelope['timestamp']?.toString() ?? '',
      )?.toUtc();
      if (authenticated) _scheduleExpiration();
      return envelope['data'];
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (authenticated && status == 401 && epoch == _sessionEpoch) {
        sessionNotice = 'Tu sesión venció. Inicia sesión nuevamente.';
        try {
          await clear();
        } catch (_) {
          // Preserve the server's authentication error after hiding private data.
        }
      }
      final payload = e.response?.data;
      final error = payload is Map && payload['error'] is Map
          ? payload['error'] as Map
          : <String, dynamic>{};
      final code = error['code']?.toString() ?? '';
      final fields = error['fields'] is Map
          ? Map<String, String>.from(error['fields'] as Map)
          : <String, String>{};
      final message =
          errorTranslations[code] ??
          error['message']?.toString() ??
          (status == 429
              ? 'Demasiadas solicitudes. Espera un momento.'
              : 'No se pudo conectar con Andaria. Comprueba tu conexión.');
      throw ApiFailure(message, code: code, fields: fields, status: status);
    }
  }

  Future<Uint8List> privateImage(String path) async {
    if (!signedIn) {
      throw const ApiFailure('Inicia sesión para ver el comprobante.');
    }
    final epoch = _sessionEpoch;
    try {
      final result = await client.get<List<int>>(
        path,
        options: Options(
          responseType: ResponseType.bytes,
          headers: {if (!kIsWeb) 'Cookie': _cookie},
        ),
      );
      if (epoch != _sessionEpoch) {
        throw const ApiFailure('La sesión cambió. Regresa para continuar.');
      }
      _scheduleExpiration();
      return Uint8List.fromList(result.data!);
    } on DioException catch (e) {
      if (e.response?.statusCode == 401 && epoch == _sessionEpoch) {
        await clear();
      }
      throw const ApiFailure('No se pudo cargar el comprobante.');
    }
  }

  String photoUrl(String type, int id, int photo) =>
      '${client.options.baseUrl}/$type/$id/photos/$photo';
  Future<void> updateProfile(Json payload) async {
    user = Map<String, dynamic>.from(
      await request('/me', method: 'PATCH', body: payload, authenticated: true)
          as Map,
    );
    notifyListeners();
  }
}
