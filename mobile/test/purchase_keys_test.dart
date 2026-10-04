import 'package:andaria_mobile/core/purchase_keys.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  test('concurrent retries obtain the same persistent purchase key', () async {
    FlutterSecureStorage.setMockInitialValues({});
    const key = 'andaria.purchase.concurrent-test';
    final values = await Future.wait([
      PurchaseKeys.obtain(key),
      PurchaseKeys.obtain(key),
    ]);
    expect(values.toSet(), hasLength(1));
    expect(await PurchaseKeys.obtain(key), values.first);
    await PurchaseKeys.accepted(key);
    expect(await PurchaseKeys.obtain(key), isNot(values.first));
  });
}
