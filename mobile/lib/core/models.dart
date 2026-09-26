import 'package:intl/intl.dart';

typedef Json = Map<String, dynamic>;
int integer(dynamic value) => value is num ? value.toInt() : 0;
List<Json> objects(dynamic value) => value is List
    ? value.whereType<Map>().map((e) => Map<String, dynamic>.from(e)).toList()
    : [];
List<String> strings(dynamic value) =>
    value is List ? value.map((e) => e.toString()).toList() : [];
String countLabel(int count, String singular, String plural) =>
    '$count ${count == 1 ? singular : plural}';
String money(int cents) =>
    'Bs ${NumberFormat('#,##0.00', 'es_BO').format(cents / 100)}';
DateTime? boliviaDate(String value) =>
    DateTime.tryParse(value)?.toUtc().subtract(const Duration(hours: 4));
String dateLabel(String value, {bool time = true}) {
  final date = boliviaDate(value);
  return date == null
      ? 'Fecha por confirmar'
      : DateFormat(
          time ? 'd MMM yyyy · HH:mm' : 'd MMM yyyy',
          'es_BO',
        ).format(date);
}

class Record {
  const Record(this.json);
  final Json json;
  String s(String key) => json[key]?.toString() ?? '';
  int n(String key) => integer(json[key]);
  bool b(String key) => json[key] == true;
  List<Json> list(String key) => objects(json[key]);
  int get id => n('id');
  String get name => s('name');
  List<int> get photos => list('photos').map((e) => integer(e['id'])).toList();
}

class TravelPackage extends Record {
  const TravelPackage(super.json);
  int get price => n('national_price_cents');
  bool get bookable => b('bookable');
  List<Record> get departures => list('departures').map(Record.new).toList();
  String get duration =>
      '${n('duration_days')} ${n('duration_days') == 1 ? 'día' : 'días'}${n('duration_nights') > 0 ? ' · ${n('duration_nights')} ${n('duration_nights') == 1 ? 'noche' : 'noches'}' : ''}';
  String get location => [
    s('agency_city'),
    s('agency_department'),
  ].where((e) => e.isNotEmpty).join(', ');
}

class Attraction extends Record {
  const Attraction(super.json);
  String get location =>
      [s('city'), s('department')].where((e) => e.isNotEmpty).join(', ');
}

const purchaseLabels = {
  'payment_review': 'Pago en revisión',
  'correction_requested': 'Corrige tu comprobante',
  'confirmed': 'Compra confirmada',
  'payment_rejected': 'Pago rechazado',
  'cancelled': 'Cancelada',
  'refund_pending': 'Reembolso pendiente',
  'refunded': 'Reembolso completado',
};

class Purchase extends Record {
  const Purchase(super.json);
  String get status => s('status');
  String get statusLabel => purchaseLabels[status] ?? 'Estado por confirmar';
}

class TravelerMinor {
  const TravelerMinor({required this.age, this.foreign = false});
  final int age;
  final bool foreign;
  Json toJson() => {'age': age, 'is_foreign': foreign};
}

class PriceBreakdown {
  PriceBreakdown({
    required int nationalAdults,
    required int foreignAdults,
    required List<TravelerMinor> minors,
    required int minimumAge,
    required int unitPrice,
    required int surcharge,
  }) {
    final paying = minors.where((e) => e.age >= minimumAge).toList();
    capacity = nationalAdults + foreignAdults + paying.length;
    travelers = nationalAdults + foreignAdults + minors.length;
    freeMinors = minors.length - paying.length;
    total =
        capacity * unitPrice +
        (foreignAdults + paying.where((e) => e.foreign).length) * surcharge;
  }
  late final int capacity, travelers, freeMinors, total;
}
