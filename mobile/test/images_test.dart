import 'dart:convert';
import 'dart:typed_data';
import 'package:andaria_mobile/core/api.dart';
import 'package:andaria_mobile/shared/images.dart';
import 'package:flutter_test/flutter_test.dart';
import 'support/gallery.dart';
import 'support/proof_fixtures.dart';

void main() {
  testWidgets('PNG and JPEG proofs preserve the original bytes', (
    tester,
  ) async {
    await tester.runAsync(() async {
      for (final sample in [
        (tinyPng, 'image/png'),
        (tinyJpg, 'image/jpeg'),
        (limitPng, 'image/png'),
      ]) {
        final bytes = base64Decode(sample.$1);
        final proof = await ProofImage.validate(bytes);
        expect(proof.mime, sample.$2);
        expect(proof.bytes, bytes);
        expect(proof.dataUrl, 'data:${sample.$2};base64,${sample.$1}');
      }
    });
  });
  testWidgets('proof dimensions reject excessive width or height', (
    tester,
  ) async {
    await tester.runAsync(() async {
      for (final sample in [widePng, tallPng]) {
        await expectLater(
          ProofImage.validate(base64Decode(sample)),
          throwsA(
            isA<ApiFailure>().having(
              (error) => error.message,
              'message',
              'Selecciona una imagen de hasta 4096 × 4096 píxeles.',
            ),
          ),
        );
      }
    });
  });
  testWidgets('receipt validation rejects damaged images before preview', (
    tester,
  ) async {
    await tester.runAsync(() async {
      final valid = await ProofImage.validate(base64Decode(tinyPng));
      expect(valid.mime, 'image/png');
      final corrupt = Uint8List.fromList([
        137,
        80,
        78,
        71,
        13,
        10,
        26,
        10,
        1,
        2,
        3,
      ]);
      await expectLater(
        ProofImage.validate(corrupt),
        throwsA(isA<ApiFailure>()),
      );
    });
  });
  test('receipt validation rejects oversized files and other formats', () {
    expect(
      () => ProofImage.fromBytes(Uint8List(5 * 1024 * 1024 + 1)),
      throwsA(isA<ApiFailure>()),
    );
    expect(
      () => ProofImage.fromBytes(Uint8List.fromList(utf8.encode('%PDF-1.7'))),
      throwsA(isA<ApiFailure>()),
    );
  });
}
