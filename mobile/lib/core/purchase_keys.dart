import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:uuid/uuid.dart';
import 'models.dart';

class PurchaseKeys {
  static const _storage = FlutterSecureStorage();
  static final _pending = <String, Future<String>>{};
  static String storageKey(String baseUrl, int userId, Json payload) {
    final fingerprint = sha256
        .convert(utf8.encode(jsonEncode(payload)))
        .toString();
    return 'andaria.purchase.$baseUrl.$userId.$fingerprint';
  }

  static Future<String> obtain(String key) => _pending.putIfAbsent(
    key,
    () => _obtain(key).whenComplete(() {
      _pending.remove(key);
    }),
  );

  static Future<String> _obtain(String key) async {
    final previous = await _storage.read(key: key);
    if (previous != null) {
      return previous;
    }
    final value = const Uuid().v4();
    await _storage.write(key: key, value: value);
    return value;
  }

  static Future<void> accepted(String key) => _storage.delete(key: key);
}
