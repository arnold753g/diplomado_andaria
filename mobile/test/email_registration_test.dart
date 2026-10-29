import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/core/theme.dart';
import 'package:andaria_mobile/features/auth.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';
import 'package:provider/provider.dart';
import 'api_test.dart' show MemoryStorage;

void main() {
  testWidgets('pending registration waits for email confirmation without login', (
    tester,
  ) async {
    final dio = Dio();
    final adapter = DioAdapter(dio: dio);
    final storage = MemoryStorage();
    final api = AndariaApi(
      baseUrl: 'https://example.test/api/v1',
      dio: dio,
      storage: storage,
    );
    addTearDown(() {
      api.dispose();
      dio.close();
    });
    adapter.onGet(
      '/auth/mobile/options',
      (server) => server.reply(200, {
        'success': true,
        'data': {'registration_enabled': true, 'google_enabled': false},
      }),
    );
    var loginRequests = 0;
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (request, handler) {
          if (request.path == '/auth/login') loginRequests++;
          handler.next(request);
        },
      ),
    );
    adapter.onPost(
      '/auth/register',
      (server) => server.reply(202, {
        'success': true,
        'data': {'verification_required': true},
      }),
      data: {
        'email': 'new@example.test',
        'password': 'security-test-password',
        'first_name': 'Nuevo',
        'last_name': 'Viajero',
      },
    );
    await tester.pumpWidget(
      ChangeNotifierProvider.value(
        value: api,
        child: MaterialApp(theme: andariaTheme, home: const AuthPage()),
      ),
    );
    await tester.pumpAndSettle();
    // Switch to registration using the public action, then fill its labelled fields.
    await tester.ensureVisible(find.text('Crear una cuenta'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Crear una cuenta'));
    await tester.pumpAndSettle();
    final fields = find.byType(TextFormField);
    await tester.enterText(fields.at(0), 'Nuevo');
    await tester.enterText(fields.at(1), 'Viajero');
    await tester.enterText(fields.at(2), 'new@example.test');
    await tester.enterText(fields.at(3), 'security-test-password');
    final submit = find.widgetWithText(
      FilledButton,
      'Crear cuenta y continuar',
    );
    await tester.ensureVisible(submit);
    await tester.pumpAndSettle();
    await tester.tap(submit);
    await tester.pumpAndSettle();
    expect(
      find.text(
        'Revisa tu correo y confirma el enlace antes de iniciar sesión.',
      ),
      findsOneWidget,
    );
    expect(loginRequests, 0);
    expect(api.signedIn, isFalse);
    expect(storage.value, isNull);
    await tester.pumpWidget(const SizedBox.shrink());
  });
}
