import 'dart:convert';
import 'support/gallery.dart';
import 'package:andaria_mobile/shared/widgets.dart';
import 'package:andaria_mobile/shared/images.dart';
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/core/models.dart';
import 'package:andaria_mobile/core/theme.dart';
import 'package:andaria_mobile/features/checkout.dart';
import 'package:andaria_mobile/features/purchases.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:provider/provider.dart';

class CheckoutApi extends AndariaApi {
  CheckoutApi({this.errorCode})
    : super(baseUrl: 'https://example.test/api/v1') {
    user = {'id': 1, 'role': 'turista'};
  }
  final String? errorCode;
  @override
  bool get signedIn => true;
  Json? sent;
  String? key;
  int submissions = 0;
  @override
  Future<dynamic> request(
    String path, {
    String method = 'GET',
    Json? body,
    Json? query,
    bool authenticated = false,
    String? idempotencyKey,
  }) async {
    if (path.endsWith('payment-options')) {
      return {
        'methods': ['transfer'],
        'minimum_paying_age': 6,
        'bank_name': 'Banco de prueba',
        'account_holder': 'Agencia de prueba',
        'account_number': '1234',
      };
    }
    if (path == '/me/purchases') {
      sent = body;
      key = idempotencyKey;
      submissions++;
      if (errorCode != null) {
        throw ApiFailure(
          errorCode == 'PURCHASE_PRICE_CHANGED'
              ? 'El precio cambió.'
              : 'La disponibilidad de la salida cambió.',
          code: errorCode!,
          status: 409,
        );
      }
      return {'id': 42};
    }
    if (path == '/me/purchases/42') {
      return {
        'id': 42,
        'reference': 'COMPRA-42',
        'package_name': 'Experiencia de prueba',
        'status': 'payment_review',
        'total_cents': 55000,
      };
    }
    throw StateError('Unexpected request: $path');
  }
}

void main() {
  for (final code in [
    null,
    'PURCHASE_PRICE_CHANGED',
    'CAPACITY_UNAVAILABLE',
    'DEPARTURE_NOT_BOOKABLE',
    'ACCOUNT_CHANGED',
  ]) {
    testWidgets('purchase confirmation and server conflict: $code', (
      tester,
    ) async {
      await initializeDateFormatting('es_BO');
      FlutterSecureStorage.setMockInitialValues({});
      final previousPicker = ImagePickerPlatform.instance;
      ImagePickerPlatform.instance = Gallery();
      await tester.runAsync(() => ProofImage.validate(base64Decode(tinyPng)));
      final api = CheckoutApi(
        errorCode: code == 'ACCOUNT_CHANGED' ? null : code,
      );
      addTearDown(() {
        ImagePickerPlatform.instance = previousPicker;
        api.dispose();
      });
      await tester.pumpWidget(
        ChangeNotifierProvider<AndariaApi>.value(
          value: api,
          child: MaterialApp(
            theme: andariaTheme,
            home: CheckoutPage(
              item: const TravelPackage({
                'id': 1,
                'name': 'Experiencia de prueba',
                'national_price_cents': 25000,
                'foreign_surcharge_cents': 5000,
              }),
              departure: const Record({
                'id': 7,
                'available_capacity': 10,
                'starts_at': '2026-11-01T12:00:00Z',
              }),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.byTooltip('Agregar Adultos extranjeros').hitTestable(),
        200,
        scrollable: find.byType(Scrollable).last,
      );
      await tester.tap(find.byTooltip('Agregar Adultos extranjeros'));
      await tester.pump();
      await tester.scrollUntilVisible(
        find.text('Seleccionar comprobante'),
        300,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.runAsync(() async {
        await tester.tap(find.text('Seleccionar comprobante'));
        await Future<void>.delayed(const Duration(milliseconds: 200));
      });
      await tester.pumpAndSettle();
      expect(find.text('Cambiar imagen'), findsOneWidget);
      final submit = find.text('Enviar compra · Bs 550,00');
      await tester.scrollUntilVisible(
        submit,
        200,
        scrollable: find.byType(Scrollable).first,
      );
      if (code == null) {
        final callback = tester
            .widget<BusyButton>(
              find.ancestor(of: submit, matching: find.byType(BusyButton)),
            )
            .onPressed!;
        callback();
        callback();
      } else {
        await tester.tap(submit);
      }
      await tester.pumpAndSettle();
      expect(find.text('Enviar compra a revisión'), findsOneWidget);
      expect(find.byType(AlertDialog, skipOffstage: false), findsOneWidget);
      expect(api.submissions, 0);
      if (code == 'ACCOUNT_CHANGED') {
        api.user = {'id': 2, 'role': 'turista'};
        api.notifyListeners();
      }
      await tester.tap(find.text('Enviar'));
      await tester.pumpAndSettle();
      if (code == 'ACCOUNT_CHANGED') {
        expect(api.submissions, 0);
        expect(api.key, isNull);
        expect(find.text('La sesión cambió'), findsOneWidget);
        expect(tester.takeException(), isNull);
        await tester.pumpWidget(const SizedBox.shrink());
        return;
      }
      expect(api.submissions, 1);
      expect(api.sent!['expected_total_cents'], 55000);
      expect(api.sent!['foreign_adults'], 1);
      expect(api.sent!['departure_id'], 7);
      expect(api.sent!['payment_proof'], startsWith('data:image/png;base64,'));
      expect(api.key, matches(RegExp(r'^[a-zA-Z0-9_-]{16,128}$')));
      if (code != null) {
        expect(find.byType(PurchaseDetailPage), findsNothing);
        final action = code == 'PURCHASE_PRICE_CHANGED'
            ? 'Revisar precio actualizado'
            : 'Revisar salidas disponibles';
        await tester.scrollUntilVisible(
          find.text(action).hitTestable(),
          200,
          scrollable: find.byType(Scrollable).first,
        );
        expect(find.text(action), findsOneWidget);
        expect(
          find.textContaining('Si ya transferiste el dinero'),
          findsOneWidget,
        );
        await tester.scrollUntilVisible(
          find.text('Enviar compra · Bs 550,00').hitTestable(),
          200,
          scrollable: find.byType(Scrollable).first,
        );
        await tester.pumpAndSettle();
        final button = tester.widget<FilledButton>(
          find.ancestor(
            of: find.text('Enviar compra · Bs 550,00'),
            matching: find.byType(FilledButton),
          ),
        );
        expect(button.onPressed, isNull);
      } else {
        expect(find.byType(PurchaseDetailPage), findsOneWidget);
        expect(find.text('Pago en revisión'), findsOneWidget);
      }
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox.shrink());
    });
  }
}
