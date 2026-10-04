import 'package:flutter/material.dart';
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/main.dart';
import 'package:andaria_mobile/features/auth.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';
import 'api_test.dart' show MemoryStorage;

void main() {
  for (final large in [false, true]) {
    testWidgets('public catalog navigation, compact large text: $large', (
      tester,
    ) async {
      if (large) {
        tester.view.physicalSize = const Size(320, 740);
        tester.view.devicePixelRatio = 1;
        tester.platformDispatcher.textScaleFactorTestValue = 1.6;
        addTearDown(() {
          tester.view.resetPhysicalSize();
          tester.view.resetDevicePixelRatio();
          tester.platformDispatcher.clearTextScaleFactorTestValue();
        });
      }
      final dio = Dio();
      final adapter = DioAdapter(dio: dio);
      final api = AndariaApi(
        baseUrl: 'https://example.test/api/v1',
        dio: dio,
        storage: MemoryStorage(),
      );
      adapter.onGet(
        '/packages',
        (server) => server.reply(200, {
          'success': true,
          'data': {
            'packages': [],
            'pagination': {'total': 0},
          },
        }),
        queryParameters: {'page': 1, 'limit': 12, 'sort': 'next_departure'},
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
        '/auth/mobile/options',
        (server) => server.reply(200, {
          'success': true,
          'data': {'registration_enabled': true, 'google_enabled': false},
        }),
      );
      await tester.pumpWidget(AndariaApp(api: api));
      await tester.pumpAndSettle();
      expect(find.text('Sal de la rutina.\nDescubre Bolivia.'), findsOneWidget);
      await tester.tap(find.text('Favoritos'));
      await tester.pumpAndSettle();
      expect(find.text('Guarda lo que te inspira'), findsOneWidget);
      await tester.tap(find.text('Mis compras'));
      await tester.pumpAndSettle();
      expect(find.text('Tus próximas experiencias'), findsOneWidget);
      await tester.tap(find.text('Perfil'));
      await tester.pumpAndSettle();
      expect(find.text('Viaja con Andaria'), findsOneWidget);
      final accessPoint = tester.getCenter(find.text('Iniciar sesión'));
      await tester.tapAt(accessPoint);
      await tester.tapAt(accessPoint);
      await tester.pumpAndSettle();
      expect(find.byType(AuthPage), findsOneWidget);
      expect(tester.takeException(), isNull);
      await tester.pumpWidget(const SizedBox.shrink());
      api.dispose();
      dio.close();
    });
  }
}
