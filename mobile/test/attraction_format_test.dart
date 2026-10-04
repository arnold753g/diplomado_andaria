import 'package:andaria_mobile/core/attraction_format.dart';
import 'package:andaria_mobile/core/models.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'opening days are ISO Monday=1 Sunday=7 and preserve overnight hours',
    () {
      final hours = attractionHours(
        const Attraction({
          'schedule_mode': 'scheduled',
          'opening_days': [7, 1],
          'opening_time': '20:00',
          'closing_time': '02:00',
        }),
      );
      expect(hours, contains('Lunes, Domingo'));
      expect(hours, contains('cierre al día siguiente'));
    },
  );
  test('all-day opening is still limited to the configured days', () {
    expect(
      attractionHours(
        const Attraction({
          'schedule_mode': 'all_day',
          'opening_days': [1],
        }),
      ),
      'Lunes · 24 horas',
    );
  });
  test('season handles year boundaries and one month', () {
    expect(
      attractionSeason(
        const Attraction({
          'season_mode': 'months',
          'season_start_month': 11,
          'season_end_month': 2,
        }),
      ),
      contains('hasta el año siguiente'),
    );
    expect(
      attractionSeason(
        const Attraction({
          'season_mode': 'months',
          'season_start_month': 3,
          'season_end_month': 3,
        }),
      ),
      'Temporada: marzo.',
    );
  });
}
