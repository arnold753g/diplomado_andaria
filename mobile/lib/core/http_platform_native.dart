import 'dart:io';
import 'package:dio/dio.dart';

void configurePlatformClient(Dio client) {}

Iterable<({String value, bool revoked})> sessionCookies(Headers headers) sync* {
  for (final header in headers['set-cookie'] ?? <String>[]) {
    final cookie = Cookie.fromSetCookieValue(header);
    if (cookie.name == 'starter_session') {
      yield (
        value: '${cookie.name}=${cookie.value}',
        revoked: cookie.value.isEmpty || cookie.maxAge == 0,
      );
    }
  }
}
