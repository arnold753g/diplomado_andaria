import 'package:andaria_mobile/core/api.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';

class MemoryStorage implements SessionStorage {
  String? value;
  bool failClear = false;
  @override
  Future<String?> read() async => value;
  @override
  Future<void> write(String value) async {
    this.value = value;
  }

  @override
  Future<void> clear() async {
    if (failClear) throw StateError('Secure store unavailable');
    value = null;
  }
}

void main() {
  late Dio dio;
  late DioAdapter adapter;
  late MemoryStorage storage;
  late AndariaApi api;
  final session = {
    'success': true,
    'data': {
      'csrf_token': 'csrf-test',
      'user': {
        'id': 2,
        'role': 'turista',
        'first_name': 'Ana',
        'last_name': 'Paz',
      },
    },
  };
  setUp(() {
    dio = Dio();
    adapter = DioAdapter(dio: dio);
    storage = MemoryStorage();
    api = AndariaApi(
      baseUrl: 'https://example.test/api/v1',
      dio: dio,
      storage: storage,
    );
  });
  tearDown(() {
    api.dispose();
    dio.close();
  });
  Future<void> login() async {
    adapter.onPost(
      '/auth/login',
      (server) => server.reply(
        200,
        session,
        headers: {
          'content-type': ['application/json'],
          'set-cookie': ['starter_session=secret; Path=/; HttpOnly; Secure'],
        },
      ),
      data: {'email': 'ana@example.test', 'password': 'long password'},
    );
    await api.login('ana@example.test', 'long password');
  }

  test(
    'login persists only the cookie and sends CSRF for protected writes',
    () async {
      await login();
      expect(api.signedIn, isTrue);
      expect(storage.value, 'starter_session=secret');
      RequestOptions? sent;
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            sent = options;
            handler.next(options);
          },
        ),
      );
      adapter.onPut(
        '/me/favorites/1',
        (server) => server.reply(200, {
          'success': true,
          'data': {'favorite': true},
        }),
      );
      await api.request('/me/favorites/1', method: 'PUT', authenticated: true);
      expect(sent!.headers['Cookie'], 'starter_session=secret');
      expect(sent!.headers['X-CSRF-Token'], 'csrf-test');
    },
  );
  test('restore retrieves the user and fresh CSRF from the server', () async {
    storage.value = 'starter_session=secret';
    adapter.onGet('/auth/session', (server) => server.reply(200, session));
    await api.restore();
    expect(api.fullName, 'Ana Paz');
    expect(api.signedIn, isTrue);
  });
  test(
    'server-confirmed password change remains successful when local cookie deletion fails',
    () async {
      await login();
      storage.failClear = true;
      final payload = {
        'current_password': 'long password',
        'new_password': 'new password more',
      };
      adapter.onPost(
        '/me/password',
        (server) => server.reply(
          200,
          {'success': true, 'data': null},
          headers: {
            'content-type': ['application/json'],
            'set-cookie': ['starter_session=; Path=/; Max-Age=0; HttpOnly'],
          },
        ),
        data: payload,
      );
      await api.request(
        '/me/password',
        method: 'POST',
        body: payload,
        authenticated: true,
      );
      expect(api.signedIn, isFalse);
      expect(api.user, isNull);
      await expectLater(
        api.request('/me/favorites', authenticated: true),
        throwsA(isA<ApiFailure>()),
      );
    },
  );
  test(
    'expired private request keeps its authentication error when deletion fails',
    () async {
      await login();
      storage.failClear = true;
      adapter.onGet(
        '/me/purchases',
        (server) => server.reply(401, {
          'success': false,
          'error': {'code': 'SESSION_EXPIRED'},
        }),
      );
      await expectLater(
        api.request('/me/purchases', authenticated: true),
        throwsA(
          isA<ApiFailure>().having((e) => e.status, 'server status', 401),
        ),
      );
      expect(api.signedIn, isFalse);
      expect(api.user, isNull);
    },
  );
  test('expired session clears all local account state', () async {
    await login();
    adapter.onGet(
      '/me/purchases',
      (server) => server.reply(401, {
        'success': false,
        'error': {'code': 'SESSION_EXPIRED'},
      }),
    );
    await expectLater(
      api.request('/me/purchases', authenticated: true),
      throwsA(isA<ApiFailure>()),
    );
    expect(api.signedIn, isFalse);
    expect(api.user, isNull);
    expect(storage.value, isNull);
  });
  test('non-tourist account is revoked and refused', () async {
    adapter.onPost(
      '/auth/login',
      (server) => server.reply(
        200,
        {
          'success': true,
          'data': {
            'csrf_token': 'csrf',
            'user': {'role': 'admin'},
          },
        },
        headers: {
          'content-type': ['application/json'],
          'set-cookie': ['starter_session=admin-secret; Path=/; HttpOnly'],
        },
      ),
      data: {'email': 'admin@example.test', 'password': 'long password'},
    );
    var revoked = false;
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (options.path == '/auth/logout') {
            revoked = options.headers['X-CSRF-Token'] == 'csrf';
          }
          handler.next(options);
        },
      ),
    );
    adapter.onPost(
      '/auth/logout',
      (server) => server.reply(200, {'success': true}),
    );
    await expectLater(
      api.login('admin@example.test', 'long password'),
      throwsA(isA<ApiFailure>()),
    );
    expect(revoked, isTrue);
    expect(storage.value, isNull);
    expect(api.signedIn, isFalse);
  });
  test(
    'failed restoration due to connection keeps credential for a later retry',
    () async {
      storage.value = 'starter_session=secret';
      adapter.onGet(
        '/auth/session',
        (server) => server.reply(503, {'success': false}),
      );
      await api.restore();
      expect(storage.value, isNotNull);
      expect(api.signedIn, isFalse);
      expect(api.sessionNotice, isNotNull);
    },
  );
  test('protected request cannot run without a session', () async {
    await expectLater(
      api.request('/me/purchases', authenticated: true),
      throwsA(isA<ApiFailure>()),
    );
  });
  test(
    'late restoration cannot revive an account after local logout',
    () async {
      storage.value = 'starter_session=secret';
      adapter.onGet(
        '/auth/session',
        (server) =>
            server.reply(200, session, delay: const Duration(milliseconds: 40)),
      );
      final restoring = api.restore();
      await Future<void>.delayed(const Duration(milliseconds: 10));
      await api.clear();
      await restoring;
      expect(api.signedIn, isFalse);
      expect(api.user, isNull);
      expect(storage.value, isNull);
    },
  );
  test('late 401 from previous session cannot clear a fresh login', () async {
    await login();
    adapter.onGet(
      '/old-request',
      (server) => server.reply(401, {
        'error': {'code': 'SESSION_EXPIRED'},
      }, delay: const Duration(milliseconds: 40)),
    );
    final pending = api.request('/old-request', authenticated: true);
    final assertion = expectLater(pending, throwsA(isA<ApiFailure>()));
    await login();
    await assertion;
    expect(api.signedIn, isTrue);
    expect(storage.value, isNotNull);
  });
  test(
    'absolute session expiry clears cached private state automatically',
    () async {
      final response = {
        'success': true,
        'timestamp': DateTime.now().toUtc().toIso8601String(),
        'data': {
          'csrf_token': 'csrf',
          'session_expires_at': DateTime.now()
              .toUtc()
              .add(const Duration(milliseconds: 250))
              .toIso8601String(),
          'idle_timeout_seconds': 1800,
          'user': {'id': 1, 'role': 'turista'},
        },
      };
      adapter.onPost(
        '/auth/login',
        (server) => server.reply(
          200,
          response,
          headers: {
            'content-type': ['application/json'],
            'set-cookie': ['starter_session=secret; HttpOnly'],
          },
        ),
        data: {'email': 'ana@example.test', 'password': 'long password'},
      );
      await api.login('ana@example.test', 'long password');
      expect(api.signedIn, isTrue);
      await Future<void>.delayed(const Duration(milliseconds: 400));
      expect(api.user, isNull);
      expect(storage.value, isNull);
    },
  );
  test(
    'expired restoration hides private state even if secure-store deletion fails',
    () async {
      storage.value = 'starter_session=expired';
      storage.failClear = true;
      adapter.onGet(
        '/auth/session',
        (server) => server.reply(401, {
          'error': {'code': 'SESSION_EXPIRED'},
        }),
      );
      await api.restore();
      expect(api.user, isNull);
      expect(api.signedIn, isFalse);
      expect(api.sessionNotice, contains('venció'));
      await expectLater(
        api.request('/me/favorites', authenticated: true),
        throwsA(isA<ApiFailure>()),
      );
    },
  );
}
