import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/features/shell.dart';
import 'package:andaria_mobile/main.dart' as app;
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:provider/provider.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  testWidgets('registration, password change and session revocation', (
    tester,
  ) async {
    const base = String.fromEnvironment('API_BASE_URL');
    expect(Uri.parse(base).port, 8081);
    final remote = Dio(BaseOptions(baseUrl: base, followRedirects: false));
    addTearDown(remote.close);
    final marker = await remote.get(
      '${Uri.parse(base).origin}/__android_fixture',
    );
    expect(marker.data, 'andaria-android-fixture-v1');
    await SecureSessionStorage(base).clear();
    await app.main();
    await tester.pumpAndSettle();
    // Keep Android's real IME from overwriting injected test edits on refocus.
    // Forms, validation, navigation and HTTP still execute in the Android app.
    binding.testTextInput.register();
    addTearDown(binding.testTextInput.unregister);
    final api = tester.element(find.byType(AppShell)).read<AndariaApi>();

    final vertical = find.byWidgetPredicate(
      (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
    );
    Future<void> press(String text) async {
      final target = find.text(text).hitTestable();
      if (target.evaluate().isEmpty && vertical.evaluate().isNotEmpty) {
        await tester.scrollUntilVisible(target, 180, scrollable: vertical.last);
      }
      await tester.tap(target);
      await tester.pumpAndSettle();
    }

    Future<void> hideKeyboard() async {
      FocusManager.instance.primaryFocus?.unfocus();
      await tester.pumpAndSettle();
    }

    Finder input(String label) => find.byWidgetPredicate(
      (w) => w is TextField && w.decoration?.labelText == label,
    );
    Future<void> enter(String label, String value) async {
      await hideKeyboard();
      tester.state<ScrollableState>(vertical.last).position.jumpTo(0);
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        input(label).hitTestable(),
        150,
        scrollable: vertical.last,
      );
      await tester.enterText(input(label), value);
      await tester.pump();
      expect(tester.widget<TextField>(input(label)).controller!.text, value);
    }

    await press('Perfil');
    await press('Iniciar sesión');
    await press('Crear una cuenta');
    await enter('Nombres', 'Registro');
    await enter('Apellidos', 'Android');
    final email =
        'android-new-${DateTime.now().microsecondsSinceEpoch}@example.test';
    await enter('Correo electrónico', email);
    await enter('Contraseña', 'corta');
    await hideKeyboard();
    await press('Crear cuenta y continuar');
    expect(
      find.textContaining('La contraseña es demasiado corta'),
      findsOneWidget,
    );
    await enter('Contraseña', 'andaria-register-old');
    await hideKeyboard();
    await press('Crear cuenta y continuar');
    for (var attempt = 0; !api.signedIn && attempt < 30; attempt++) {
      await tester.pump(const Duration(milliseconds: 200));
    }
    await tester.pumpAndSettle();
    expect(
      api.signedIn,
      isTrue,
      reason: tester
          .widgetList<Text>(find.byType(Text))
          .map((w) => w.data ?? '')
          .join(' | '),
    );
    expect(api.user!['role'], 'turista');
    expect(api.fullName, 'Registro Android');

    final second = await remote.post(
      '/auth/login',
      data: {'email': email, 'password': 'andaria-register-old'},
    );
    final cookie = second.headers['set-cookie']!.first.split(';').first;
    remote.options.headers['Cookie'] = cookie;
    expect((await remote.get('/auth/session')).statusCode, 200);
    await press('Cambiar contraseña');
    await enter('Contraseña actual', 'incorrecta');
    await enter('Nueva contraseña', 'andaria-register-new');
    await enter('Repite la nueva contraseña', 'andaria-register-new');
    await hideKeyboard();
    await press('Guardar contraseña');
    expect(
      api.signedIn,
      isTrue,
      reason: 'Incorrect old password must not close the session.',
    );
    await enter('Contraseña actual', 'andaria-register-old');
    await hideKeyboard();
    await press('Guardar contraseña');
    expect(api.signedIn, isFalse);
    await expectLater(
      remote.get('/auth/session'),
      throwsA(
        isA<DioException>().having(
          (e) => e.response?.statusCode,
          'revoked session',
          401,
        ),
      ),
    );
    await press('Iniciar sesión');
    await enter('Correo electrónico', email);
    await enter('Contraseña', 'andaria-register-old');
    await hideKeyboard();
    await press('Iniciar sesión');
    expect(
      find.text('El correo o la contraseña no son correctos.'),
      findsOneWidget,
    );
    expect(api.signedIn, isFalse);
    await enter('Contraseña', 'andaria-register-new');
    await hideKeyboard();
    await press('Iniciar sesión');
    expect(api.signedIn, isTrue);
    await binding.convertFlutterSurfaceToImage();
    await tester.pump();
    await binding.takeScreenshot('08-perfil');
    await api.logout();
    await tester.pumpAndSettle();
    expect(api.signedIn, isFalse);
    expect(tester.takeException(), isNull);
  });
}
