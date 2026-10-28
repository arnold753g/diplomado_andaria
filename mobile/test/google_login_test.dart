import 'package:andaria_mobile/core/api.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';
import 'api_test.dart' show MemoryStorage;

void main() {
  late Dio dio;
  late DioAdapter adapter;
  late MemoryStorage storage;
  late AndariaApi api;

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

  tearDown(() => api.dispose());

  test('Google identity becomes an Andaria cookie session with CSRF', () async {
    adapter.onPost(
      '/auth/google/mobile',
      (server) => server.reply(
        200,
        {
          'success': true,
          'data': {
            'csrf_token': 'andaria-csrf',
            'user': {'id': 42, 'role': 'turista', 'google_linked': true},
          },
        },
        headers: {
          'content-type': ['application/json'],
          'set-cookie': ['starter_session=andaria-cookie; HttpOnly; Secure'],
        },
      ),
      data: {'id_token': 'google-signed-id-token'},
    );

    await api.loginGoogle('google-signed-id-token');
    expect(api.signedIn, isTrue);
    expect(api.user!['id'], 42);
    expect(api.user!['google_linked'], isTrue);
    expect(storage.value, 'starter_session=andaria-cookie');
    expect(storage.value, isNot(contains('google-signed-id-token')));

    RequestOptions? protectedRequest;
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          protectedRequest = options;
          handler.next(options);
        },
      ),
    );
    adapter.onPut(
      '/me/favorites/1',
      (server) => server.reply(200, {'success': true, 'data': null}),
    );
    await api.request('/me/favorites/1', method: 'PUT', authenticated: true);
    expect(protectedRequest!.headers['Cookie'], storage.value);
    expect(protectedRequest!.headers['X-CSRF-Token'], 'andaria-csrf');
    expect(protectedRequest!.headers.containsKey('Authorization'), isFalse);
  });

  test(
    'rejected Google login clears the previous account and cookie',
    () async {
      storage.value = 'starter_session=old-account';
      adapter.onGet(
        '/auth/session',
        (server) => server.reply(200, {
          'success': true,
          'data': {
            'csrf_token': 'old-csrf',
            'user': {'id': 1, 'role': 'turista'},
          },
        }),
      );
      await api.restore();
      expect(api.signedIn, isTrue);
      adapter.onPost(
        '/auth/google/mobile',
        (server) => server.reply(401, {
          'success': false,
          'error': {
            'code': 'GOOGLE_TOKEN_INVALID',
            'message': 'Token rechazado',
          },
        }),
        data: {'id_token': 'invalid-token'},
      );
      await expectLater(
        api.loginGoogle('invalid-token'),
        throwsA(
          isA<ApiFailure>().having(
            (e) => e.code,
            'code',
            'GOOGLE_TOKEN_INVALID',
          ),
        ),
      );
      expect(api.user, isNull);
      expect(api.signedIn, isFalse);
      expect(storage.value, isNull);
    },
  );

  test(
    'existing local email reports explicit linking without creating a session',
    () async {
      const message = 'Ingresa con contraseña y vincula Google desde la web';
      adapter.onPost(
        '/auth/google/mobile',
        (server) => server.reply(409, {
          'success': false,
          'error': {'code': 'GOOGLE_EMAIL_EXISTS', 'message': message},
        }),
        data: {'id_token': 'verified-google-identity'},
      );
      await expectLater(
        api.loginGoogle('verified-google-identity'),
        throwsA(
          isA<ApiFailure>()
              .having((e) => e.code, 'code', 'GOOGLE_EMAIL_EXISTS')
              .having((e) => e.message, 'message', message),
        ),
      );
      expect(api.signedIn, isFalse);
      expect(storage.value, isNull);
    },
  );
}
