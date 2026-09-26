import 'models.dart';

String attractionHours(Attraction item) {
  if (item.s('schedule_mode') == 'unspecified') {
    return item.s('opening_hours').isEmpty
        ? 'Consulta los horarios antes de tu visita.'
        : item.s('opening_hours');
  }
  const names = [
    'Lunes',
    'Martes',
    'Miércoles',
    'Jueves',
    'Viernes',
    'Sábado',
    'Domingo',
  ];
  final days =
      (item.json['opening_days'] as List? ?? [])
          .whereType<int>()
          .where((day) => day >= 1 && day <= 7)
          .toList()
        ..sort();
  final label = days.length == 7
      ? 'Todos los días'
      : days.map((day) => names[day - 1]).join(', ');
  if (item.s('schedule_mode') == 'all_day') return '$label · 24 horas';
  if (item.s('schedule_mode') != 'scheduled') {
    return 'Consulta los horarios antes de tu visita.';
  }
  final overnight = item.s('closing_time').compareTo(item.s('opening_time')) < 0
      ? ' (cierre al día siguiente)'
      : '';
  return '$label · ${item.s('opening_time')}–${item.s('closing_time')}$overnight (Bolivia)';
}

String attractionSeason(Attraction item) {
  if (item.s('season_mode') == 'all_year') return 'Disponible todo el año.';
  if (item.s('season_mode') != 'months') return 'Temporada por confirmar.';
  const months = [
    'enero',
    'febrero',
    'marzo',
    'abril',
    'mayo',
    'junio',
    'julio',
    'agosto',
    'septiembre',
    'octubre',
    'noviembre',
    'diciembre',
  ];
  final start = item.n('season_start_month'), end = item.n('season_end_month');
  if (start < 1 || start > 12 || end < 1 || end > 12) {
    return 'Temporada por confirmar.';
  }
  if (start == end) return 'Temporada: ${months[start - 1]}.';
  return 'Temporada: ${months[start - 1]} a ${months[end - 1]}${end < start ? ' (hasta el año siguiente)' : ''}.';
}
