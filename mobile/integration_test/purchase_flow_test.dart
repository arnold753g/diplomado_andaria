import 'dart:io';
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/features/catalog.dart';
import 'package:andaria_mobile/features/shell.dart';
import 'package:andaria_mobile/main.dart' as app;
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:integration_test/integration_test.dart';
import 'package:provider/provider.dart';
import '../test/support/gallery.dart';

// This test requires TestAndroidFixture on port 8081. All accounts and purchases
// belong to its temporary schema. The simulated gallery supplies a test PNG.
void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  testWidgets('Android purchase, correction, confirmation and cancellation', (
    tester,
  ) async {
    const base = String.fromEnvironment('API_BASE_URL');
    expect(
      Uri.parse(base).port,
      8081,
      reason: 'Run only against the isolated Android fixture.',
    );
    final marker = await Dio().get(
      '${Uri.parse(base).origin}/__android_fixture',
    );
    expect(marker.data, 'andaria-android-fixture-v1');
    await SecureSessionStorage(base).clear();
    ImagePickerPlatform.instance = Gallery();
    await app.main();
    await tester.pumpAndSettle();
    await tester.tap(find.text('Perfil'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Iniciar sesión'));
    await tester.pumpAndSettle();
    await tester.enterText(
      find.byType(TextFormField).at(0),
      'android-tourist@example.test',
    );
    await tester.enterText(
      find.byType(TextFormField).at(1),
      'andaria-android-test',
    );
    await tester.tap(find.text('Iniciar sesión'));
    await tester.pumpAndSettle();
    expect(find.text('Android Prueba'), findsOneWidget);
    final api = tester.element(find.byType(AppShell)).read<AndariaApi>();
    expect(api.signedIn, isTrue);
    await api.request('/me/favorites/1', method: 'DELETE', authenticated: true);
    await tester.tap(find.text('Mis datos'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextFormField).at(2), '71112223');
    FocusManager.instance.primaryFocus?.unfocus();
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Guardar cambios').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Guardar cambios'));
    await tester.pumpAndSettle();
    expect(api.user!['phone'], '71112223');
    await tester.tap(find.text('Explorar'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Lugares'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byType(AttractionCard).hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.byType(AttractionCard));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Guardar en favoritos'));
    await tester.pumpAndSettle();
    expect(find.byTooltip('Quitar de favoritos'), findsOneWidget);
    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Favoritos'));
    await tester.pumpAndSettle();
    expect(find.text('Mirador Android de prueba'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Quitar de favoritos').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Quitar de favoritos'));
    await tester.pumpAndSettle();
    expect(find.text('Tu lista empieza con un lugar'), findsOneWidget);
    await tester.tap(find.text('Explorar'));
    await tester.pumpAndSettle();
    await tester.drag(
      find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
      const Offset(0, 1200),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Experiencias'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byType(PackageCard),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.drag(find.byType(Scrollable).first, const Offset(0, -240));
    await tester.pumpAndSettle();
    await tester.tap(find.byType(PackageCard).first);
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Elegir esta salida').hitTestable(),
      300,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Elegir esta salida'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.byTooltip('Agregar Adultos extranjeros').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.byTooltip('Agregar Adultos extranjeros'));
    await tester.pump();
    await tester.scrollUntilVisible(
      find.text('Agregar menor').hitTestable(),
      150,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Agregar menor'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Agregar'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Seleccionar comprobante').hitTestable(),
      300,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.scrollUntilVisible(
      find.text('Compartir QR para pagar').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    expect(find.text('Compartir QR para pagar'), findsOneWidget);
    await tester.ensureVisible(find.text('Transferencia'));
    await tester.tap(find.text('Transferencia'));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Seleccionar comprobante').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Seleccionar comprobante'));
    await tester.pumpAndSettle(const Duration(milliseconds: 300));
    expect(find.text('Cambiar imagen'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Revisar imagen').hitTestable(),
      150,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Revisar imagen'));
    await tester.pumpAndSettle();
    expect(find.text('Revisa tu comprobante'), findsOneWidget);
    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();
    await binding.convertFlutterSurfaceToImage();
    await tester.pump();
    await binding.takeScreenshot('04-compra');
    await tester.scrollUntilVisible(
      find.text('Enviar compra · Bs 550,00'),
      150,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Enviar compra · Bs 550,00'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Enviar'));
    await tester.pumpAndSettle();
    expect(find.text('Pago en revisión'), findsOneWidget);

    final purchases = await api.request('/me/purchases', authenticated: true);
    final purchase = purchases['purchases'][0] as Map;
    final id = purchase['id'];
    expect(purchase['total_cents'], 55000);
    expect(purchase['capacity_count'], 2);
    expect(purchase['free_minor_count'], 1);
    final agency = Dio(BaseOptions(baseUrl: base));
    final login = await agency.post(
      '/auth/login',
      data: {
        'email': 'android-agency@example.test',
        'password': 'andaria-android-test',
      },
    );
    final cookie = Cookie.fromSetCookieValue(
      login.headers['set-cookie']!.first,
    );
    agency.options.headers = {
      'Cookie': '${cookie.name}=${cookie.value}',
      'X-CSRF-Token': login.data['data']['csrf_token'],
    };
    await agency.patch(
      '/agency/purchases/$id/review',
      data: {
        'version': purchase['version'],
        'decision': 'request_correction',
        'reason': 'Adjunta otro comprobante de prueba.',
      },
    );
    Future<void> refresh() async {
      await tester.drag(find.byType(ListView).last, const Offset(0, 1000));
      await tester.pumpAndSettle();
      await tester.drag(find.byType(ListView).last, const Offset(0, 600));
      await tester.pumpAndSettle();
    }

    await refresh();
    expect(find.text('Corrige tu comprobante'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Corregir comprobante'),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Corregir comprobante'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Seleccionar comprobante'));
    await tester.pumpAndSettle(const Duration(milliseconds: 300));
    await tester.tap(find.text('Enviar corrección'));
    await tester.pumpAndSettle();
    expect(find.text('Pago en revisión'), findsOneWidget);
    final corrected = await api.request(
      '/me/purchases/$id',
      authenticated: true,
    );
    expect(corrected['agency_phone'], '70000000');
    expect(corrected['agency_email'], 'android-agency@example.test');
    await agency.patch(
      '/agency/purchases/$id/review',
      data: {'version': corrected['version'], 'decision': 'confirm'},
    );
    await refresh();
    expect(find.text('Compra confirmada'), findsOneWidget);
    await binding.takeScreenshot('05-confirmacion');
    await tester.scrollUntilVisible(
      find.text('Cancelar compra'),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Cancelar compra'));
    await tester.pumpAndSettle();
    final fields = find.byType(TextFormField);
    await tester.enterText(fields.at(0), 'Cancelación de prueba Android.');
    await tester.enterText(fields.at(1), 'Banco de prueba');
    await tester.enterText(fields.at(2), 'Android Prueba');
    await tester.enterText(fields.at(3), 'PRUEBA-456');
    await tester.scrollUntilVisible(
      find.text('Confirmar cancelación'),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Confirmar cancelación'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Confirmar cancelación').last);
    await tester.pumpAndSettle();
    expect(find.text('Reembolso pendiente'), findsOneWidget);
    await binding.takeScreenshot('06-devolucion');
    final cancelled = await api.request(
      '/me/purchases/$id',
      authenticated: true,
    );
    expect(cancelled['refund_account_number'], 'PRUEBA-456');
    expect(cancelled['status'], 'refund_pending');
    await agency.patch(
      '/agency/purchases/$id/refund',
      data: {
        'version': cancelled['version'],
        'refund_reference': 'ANDROID-TEST-RETURN',
        'refund_proof': 'data:image/png;base64,$tinyPng',
      },
    );
    await refresh();
    expect(find.text('Reembolso completado'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Ver comprobante de devolución').hitTestable(),
      200,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    await tester.tap(find.text('Ver comprobante de devolución'));
    await tester.pumpAndSettle();
    expect(find.text('Comprobante de devolución'), findsOneWidget);
    await tester.tap(find.byType(BackButton));
    await tester.pumpAndSettle();
    await tester.scrollUntilVisible(
      find.text('Contacta a la agencia').hitTestable(),
      150,
      scrollable: find
          .byWidgetPredicate(
            (w) => w is Scrollable && w.axisDirection == AxisDirection.down,
          )
          .last,
    );
    expect(find.text('Abrir teléfono'), findsOneWidget);
    expect(find.text('Redactar correo'), findsOneWidget);
    await binding.takeScreenshot('09-contacto');
    await api.logout();
    await tester.pumpAndSettle();
    expect(find.text('Inicia sesión nuevamente'), findsOneWidget);
    expect(tester.takeException(), isNull);
    agency.close();
  });
}
