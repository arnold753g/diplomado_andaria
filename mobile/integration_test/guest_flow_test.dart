import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/features/catalog.dart';
import 'package:andaria_mobile/shared/widgets.dart';
import 'package:andaria_mobile/shared/photo_gallery.dart';
import 'package:andaria_mobile/main.dart' as app;
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  testWidgets('Android guest explores the real backend and opens login', (
    tester,
  ) async {
    await app.main();
    await tester.pumpAndSettle(const Duration(milliseconds: 300));
    expect(find.text('Sal de la rutina.\nDescubre Bolivia.'), findsOneWidget);
    await binding.convertFlutterSurfaceToImage();
    await tester.pump();
    await binding.takeScreenshot('01-explorar');
    final api = AndariaApi();
    final result = await api.request('/packages', query: {'limit': 1});
    final packages = result['packages'] as List;
    if (packages.isNotEmpty) {
      final name = packages.first['name'].toString();
      await tester.scrollUntilVisible(
        find.text(name),
        200,
        scrollable: find.byType(Scrollable).first,
      );
      await tester.drag(find.byType(Scrollable).first, const Offset(0, -240));
      await tester.pumpAndSettle();
      await tester.tap(find.byType(PackageCard).first);
      await tester.pumpAndSettle();
      await tester.pumpAndSettle(const Duration(milliseconds: 300));
      expect(find.text('La experiencia'), findsOneWidget);
      await binding.takeScreenshot('02-experiencia');
      await tester.tap(find.byType(TravelPhoto).last);
      await tester.pumpAndSettle();
      expect(find.byType(PhotoGalleryPage), findsOneWidget);
      await binding.takeScreenshot('07-galeria');
      await tester.tap(find.byType(BackButton).last);
      await tester.pumpAndSettle();
      await tester.tap(find.byType(BackButton));
      await tester.pumpAndSettle();
    }
    await tester.tap(find.text('Favoritos'));
    await tester.pumpAndSettle();
    expect(find.text('Guarda lo que te inspira'), findsOneWidget);
    await tester.tap(find.text('Iniciar sesión'));
    await tester.pumpAndSettle();
    expect(find.text('Correo electrónico'), findsOneWidget);
    await binding.takeScreenshot('03-acceso');
    expect(tester.takeException(), isNull);
    api.dispose();
  });
}
