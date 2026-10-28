import 'dart:async';
import 'dart:io';
import 'dart:ui' as ui;
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/core/theme.dart';
import 'package:andaria_mobile/main.dart';
import 'package:andaria_mobile/shared/brand.dart';
import 'package:andaria_mobile/shared/startup.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';
import 'api_test.dart' show MemoryStorage;

const captureHome = bool.fromEnvironment('CAPTURE_HOME');
final previewKey = GlobalKey();

Future<void> capture(WidgetTester tester, String name) async {
  if (!captureHome) return;
  final boundary =
      previewKey.currentContext!.findRenderObject() as RenderRepaintBoundary;
  await tester.runAsync(() async {
    final image = await boundary.toImage(pixelRatio: 2);
    final bytes = await image.toByteData(format: ui.ImageByteFormat.png);
    final folder = Directory('../.local/mobile-home-2026-10-23');
    await folder.create(recursive: true);
    await File(
      '${folder.path}/$name.png',
    ).writeAsBytes(bytes!.buffer.asUint8List());
    image.dispose();
  });
}

void main() {
  setUpAll(() async {
    if (captureHome) {
      await (FontLoader(
        'Outfit',
      )..addFont(rootBundle.load('assets/fonts/Outfit.ttf'))).load();
      await (FontLoader(
        'MaterialIcons',
      )..addFont(rootBundle.load('fonts/MaterialIcons-Regular.otf'))).load();
    }
  });

  testWidgets('startup follows readiness without a fixed splash delay', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(390, 844);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });
    final ready = Completer<void>();
    var calls = 0;
    await tester.pumpWidget(
      RepaintBoundary(
        key: previewKey,
        child: MaterialApp(
          theme: andariaTheme,
          debugShowCheckedModeBanner: false,
          home: AndariaStartup(
            initialize: () {
              calls++;
              return ready.future;
            },
            child: const Scaffold(body: Text('Catálogo listo')),
          ),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 400));
    expect(find.byType(AndariaLoading), findsOneWidget);
    expect(find.text('Catálogo listo'), findsNothing);
    if (captureHome) {
      await tester.runAsync(
        () => precacheImage(
          const AssetImage('assets/branding/andaria-symbol.png'),
          tester.element(find.byType(AndariaLoading)),
        ),
      );
      await tester.pump();
    }
    await capture(tester, 'carga');
    ready.complete();
    await tester.pumpAndSettle();
    expect(find.text('Catálogo listo'), findsOneWidget);
    expect(find.byType(AndariaLoading), findsNothing);
    expect(calls, 1);
  });

  testWidgets('startup error offers retry and shows loading during retry', (
    tester,
  ) async {
    var ready = Completer<void>();
    await tester.pumpWidget(
      MaterialApp(
        home: AndariaStartup(
          initialize: () => ready.future,
          child: const Text('Listo'),
        ),
      ),
    );
    ready.completeError(StateError('Initialization failed'));
    await tester.pumpAndSettle();
    expect(find.text('No pudimos iniciar la aplicación'), findsOneWidget);
    expect(find.byType(AndariaLoading), findsNothing);
    ready = Completer<void>();
    await tester.tap(find.text('Volver a intentar'));
    await tester.pump();
    expect(find.byType(AndariaLoading), findsOneWidget);
    ready.complete();
    await tester.pumpAndSettle();
    expect(find.text('Listo'), findsOneWidget);
  });

  testWidgets('brand loader respects reduced motion', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: MediaQuery(
          data: MediaQueryData(disableAnimations: true),
          child: Scaffold(body: AndariaLoading()),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(AndariaLogo), findsOneWidget);
    expect(tester.binding.hasScheduledFrame, isFalse);
    expect(tester.takeException(), isNull);
  });

  for (final size in [
    const Size(390, 844),
    const Size(320, 740),
    const Size(1000, 900),
  ]) {
    testWidgets('home and catalog modules work at ${size.width}', (
      tester,
    ) async {
      tester.view.physicalSize = size;
      tester.view.devicePixelRatio = 1;
      if (size.width == 320) {
        tester.platformDispatcher.textScaleFactorTestValue = 1.6;
      }
      addTearDown(() {
        tester.view.resetPhysicalSize();
        tester.view.resetDevicePixelRatio();
        tester.platformDispatcher.clearTextScaleFactorTestValue();
      });
      final dio = Dio();
      final adapter = DioAdapter(dio: dio);
      final api = AndariaApi(
        baseUrl: 'https://example.test/api/v1',
        dio: dio,
        storage: MemoryStorage(),
      );
      adapter.onGet(
        '/packages/options',
        (server) => server.reply(200, {
          'success': true,
          'data': {
            'departments': ['Tarija'],
          },
        }),
      );
      adapter.onGet(
        '/packages',
        (server) => server.reply(200, {
          'success': true,
          'data': {
            'packages': [
              {
                'id': 1,
                'name': 'Un recorrido para descubrir Tarija',
                'agency_city': 'Cercado',
                'agency_department': 'Tarija',
                'agency_name': 'Agencia de ejemplo',
                'duration_days': 1,
                'national_price_cents': 25000,
                'bookable': true,
              },
            ],
            'pagination': {'total': 1},
          },
        }),
        queryParameters: {'page': 1, 'limit': 12, 'sort': 'next_departure'},
      );
      var attractionsRequests = 0;
      adapter.onGet('/attractions', (server) {
        attractionsRequests++;
        server.reply(200, {
          'success': true,
          'data': {
            'attractions': [],
            'pagination': {'total': 0},
          },
        });
      }, queryParameters: {'page': 1, 'limit': 12});
      await tester.pumpWidget(
        RepaintBoundary(
          key: previewKey,
          child: AndariaApp(api: api),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(AndariaLoading), findsNothing);
      expect(find.byType(AndariaBrand), findsOneWidget);
      expect(find.text('Inicio'), findsOneWidget);
      await capture(tester, 'inicio-${size.width.toInt()}');
      await tester.scrollUntilVisible(
        find.text('Paquetes turísticos'),
        180,
        scrollable: find.byType(Scrollable).first,
      );
      expect(find.text('Paquetes turísticos'), findsOneWidget);
      expect(find.text('Atracciones turísticas'), findsOneWidget);
      await tester.ensureVisible(find.text('Atracciones turísticas'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Atracciones turísticas'));
      await tester.pumpAndSettle();
      expect(attractionsRequests, 1);
      expect(find.byTooltip('Ordenar paquetes'), findsNothing);
      await tester.ensureVisible(find.text('Paquetes turísticos'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Paquetes turísticos'));
      await tester.pumpAndSettle();
      expect(find.byTooltip('Ordenar paquetes'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox.shrink());
      api.dispose();
      dio.close();
    });
  }
}
