import 'package:dio/dio.dart';
import 'package:dio/browser.dart';

void configurePlatformClient(Dio client) {
  client.httpClientAdapter = BrowserHttpClientAdapter(withCredentials: true);
}

// HttpOnly cookies belong to the browser and cannot be read by Dart/JavaScript.
Iterable<({String value, bool revoked})> sessionCookies(Headers headers) => [];
