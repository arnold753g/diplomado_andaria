@TestOn('browser')
library;

import 'package:andaria_mobile/core/api.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';

void main() {
  late AndariaApi api;
  late DioAdapter adapter;
  final session = {
    'success': true,
    'data': {
      'csrf_token': 'browser-csrf',
      'user': {'id': 42, 'role': 'turista', 'first_name': 'Ana'},
    },
  };
  setUp(() {
    final dio = Dio();
    adapter = DioAdapter(dio: dio);
    api = AndariaApi(dio: dio);
  });
  tearDown(() => api.dispose());

  test('restores browser session without reading an HttpOnly cookie', () async {
    expect(api.client.options.baseUrl, 'http://localhost:8080/api/v1');
    adapter.onGet('/auth/session', (server) => server.reply(200, session));
    await api.restore();
    expect(api.signedIn, isTrue);
    expect(api.user?['id'], 42);
  });

  test('login uses CSRF and logout revokes the browser session', () async {
    adapter.onPost(
      '/auth/login',
      (server) => server.reply(200, session),
      data: {'email': 'ana@example.test', 'password': 'long password'},
    );
    await api.login('ana@example.test', 'long password');
    expect(api.signedIn, isTrue);
    adapter.onPatch(
      '/me',
      (server) => server.reply(200, {'success': true, 'data': {}}),
      headers: {'X-CSRF-Token': 'browser-csrf'},
    );
    await api.request('/me', method: 'PATCH', authenticated: true);
    adapter.onPost(
      '/auth/logout',
      (server) => server.reply(200, {'success': true, 'data': {}}),
      headers: {'X-CSRF-Token': 'browser-csrf'},
    );
    await api.logout();
    expect(api.signedIn, isFalse);
    expect(api.user, isNull);
  });

  test('a browser guest does not receive a session-expired notice', () async {
    adapter.onGet('/auth/session', (server) => server.reply(401, {}));
    await api.restore();
    expect(api.signedIn, isFalse);
    expect(api.sessionNotice, isNull);
  });
}
