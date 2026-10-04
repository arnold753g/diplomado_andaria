import 'package:andaria_mobile/core/password_policy.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'password limits support Spanish text without exceeding bcrypt input',
    () {
      expect(passwordError('abcdefghijkl'), isNull);
      expect(passwordError('ñ' * 36), isNull);
      expect(passwordError('ñ' * 37), contains('demasiado larga'));
      expect(passwordError('ñ' * 5), contains('demasiado corta'));
      expect(passwordError(' abcdefghijkl'), contains('espacios'));
      expect(passwordError('abcdefghijkl\u00a0'), contains('espacios'));
    },
  );
}
