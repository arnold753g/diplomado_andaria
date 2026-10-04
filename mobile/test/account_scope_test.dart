import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/shared/account_scope.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http_mock_adapter/http_mock_adapter.dart';
import 'package:provider/provider.dart';
import 'api_test.dart' show MemoryStorage;

void main() {
  testWidgets('private screen hides data when a different tourist logs in', (
    tester,
  ) async {
    final dio = Dio();
    final adapter = DioAdapter(dio: dio);
    final api = AndariaApi(
      baseUrl: 'https://example.test/api/v1',
      dio: dio,
      storage: MemoryStorage(),
    );
    Future<void> login(int user) async {
      adapter.onPost(
        '/auth/login',
        (server) => server.reply(
          200,
          {
            'success': true,
            'data': {
              'csrf_token': 'csrf',
              'user': {'id': user, 'role': 'turista'},
            },
          },
          headers: {
            'content-type': ['application/json'],
            'set-cookie': ['starter_session=secret; HttpOnly'],
          },
        ),
        data: {'email': '$user@example.test', 'password': 'long password'},
      );
      await api.login('$user@example.test', 'long password');
    }

    await tester.runAsync(() => login(1));
    final host = ChangeNotifierProvider.value(
      value: api,
      child: const MaterialApp(
        home: Scaffold(body: AccountScope(child: Text('Documento privado'))),
      ),
    );
    await tester.pumpWidget(host);
    expect(find.text('Documento privado'), findsOneWidget);
    await api.clear();
    await tester.pump();
    expect(find.text('Documento privado'), findsNothing);
    await tester.runAsync(() => login(2));
    await tester.pump();
    expect(find.text('Documento privado'), findsNothing);
    expect(find.text('La sesión cambió'), findsOneWidget);
    await tester.pumpWidget(const SizedBox.shrink());
    api.dispose();
    dio.close();
  });
}
