import 'dart:typed_data';
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/core/theme.dart';
import 'package:andaria_mobile/features/purchases.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:printing/printing.dart';
import 'package:provider/provider.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  testWidgets(
    'Android downloads, previews and rasterizes the shared purchase PDF',
    (tester) async {
      const base = String.fromEnvironment('API_BASE_URL');
      const password = String.fromEnvironment('PDF_QA_PASSWORD');
      expect(Uri.parse(base).port, 18886);
      expect(
        (await Dio().get('${Uri.parse(base).origin}/__payments_fixture')).data,
        'andaria-payments-fixture-v1',
      );
      final api = AndariaApi(baseUrl: base);
      await api.login('payments-2@example.test', password);
      await initializeDateFormatting('es_BO');
      await tester.pumpWidget(
        ChangeNotifierProvider.value(
          value: api,
          child: MaterialApp(
            theme: andariaTheme,
            home: const PurchaseDetailPage(id: 1),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(
        find.text('Ver comprobante PDF'),
        150,
        scrollable: find.byType(Scrollable).last,
      );
      await tester.tap(find.text('Ver comprobante PDF'));
      await tester.pumpAndSettle();
      expect(find.byType(PdfPreview), findsOneWidget);
      final bytes = await api.privatePdf('/me/purchases/1/receipt');
      expect(bytes.sublist(0, 5), [37, 80, 68, 70, 45]);
      // Exercise the native Android PdfRenderer used by the actual preview.
      final pages = await Printing.raster(bytes, pages: [0], dpi: 72).toList();
      expect(pages, hasLength(1));
      final Uint8List png = await pages.first.toPng();
      expect(png.length, greaterThan(1000));
      expect(pages.first.width, greaterThan(500));
      await binding.convertFlutterSurfaceToImage();
      await tester.pump();
      await binding.takeScreenshot('comprobante-pdf-android');
      await api.clear();
      await tester.pumpAndSettle();
      expect(find.byType(PdfPreview), findsNothing);
      expect(find.text('La sesión cambió'), findsOneWidget);
      api.dispose();
    },
  );
}
