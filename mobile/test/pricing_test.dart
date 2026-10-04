import 'package:andaria_mobile/core/models.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';

void main() {
  test('mixed group uses each traveler nationality and minimum paying age', () {
    final price = PriceBreakdown(
      nationalAdults: 1,
      foreignAdults: 1,
      minors: const [
        TravelerMinor(age: 5, foreign: true),
        TravelerMinor(age: 6),
        TravelerMinor(age: 8, foreign: true),
      ],
      minimumAge: 6,
      unitPrice: 25000,
      surcharge: 5000,
    );
    expect(price.travelers, 5);
    expect(price.capacity, 4);
    expect(price.freeMinors, 1);
    expect(price.total, 110000);
  });
  test('Bolivia time conversion does not depend on device time zone', () async {
    await initializeDateFormatting('es_BO');
    final shifted = boliviaDate('2026-10-08T02:30:00Z')!;
    expect(shifted.day, 7);
    expect(shifted.hour, 22);
    expect(dateLabel('2026-10-08T02:30:00Z'), contains('22:30'));
  });
}
