import '../shared/account_scope.dart';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/purchase_keys.dart';
import '../core/api.dart';
import '../core/models.dart';
import '../shared/widgets.dart';
import '../shared/images.dart';
import 'purchases.dart';
import 'profile.dart';
import '../shared/agency_contact.dart';

class CheckoutPage extends StatefulWidget {
  const CheckoutPage({super.key, required this.item, required this.departure});
  final TravelPackage item;
  final Record departure;
  @override
  State<CheckoutPage> createState() => _CheckoutPageState();
}

class _CheckoutPageState extends State<CheckoutPage> {
  int nationals = 1, foreigners = 0;
  final List<TravelerMinor> minors = [];
  String method = '';
  ProofImage? proof;
  bool busy = false,
      confirming = false,
      priceChanged = false,
      availabilityChanged = false;
  String? error;
  Json? options;
  @override
  void initState() {
    super.initState();
    loadOptions();
  }

  Future<void> loadOptions() async {
    setState(() {
      error = null;
    });
    try {
      final data = Map<String, dynamic>.from(
        await context.read<AndariaApi>().request(
              '/me/packages/${widget.item.id}/payment-options',
              authenticated: true,
            )
            as Map,
      );
      if (mounted) {
        setState(() {
          options = data;
          method = strings(data['methods']).firstOrNull ?? '';
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e.toString();
        });
      }
    }
  }

  PriceBreakdown get price => PriceBreakdown(
    nationalAdults: nationals,
    foreignAdults: foreigners,
    minors: minors,
    minimumAge: integer(
      options?['minimum_paying_age'] ?? widget.item.n('minimum_paying_age'),
    ),
    unitPrice: widget.item.price,
    surcharge: widget.item.n('foreign_surcharge_cents'),
  );
  Future<void> pick() async {
    try {
      final value = await ProofImage.pick();
      if (mounted && value != null) {
        setState(() {
          proof = value;
        });
      }
    } catch (e) {
      if (mounted) message(context, e);
    }
  }

  Future<void> addMinor() async {
    int age = 5;
    bool foreign = false;
    final minor = await showDialog<TravelerMinor>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, update) => AlertDialog(
          title: const Text('Agregar menor'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<int>(
                initialValue: age,
                decoration: const InputDecoration(labelText: 'Edad'),
                items: List.generate(
                  18,
                  (i) => DropdownMenuItem(
                    value: i,
                    child: Text(countLabel(i, 'año', 'años')),
                  ),
                ),
                onChanged: (v) => update(() {
                  age = v!;
                }),
              ),
              const SizedBox(height: 8),
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Extranjero'),
                value: foreign,
                onChanged: (v) => update(() {
                  foreign = v;
                }),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Volver'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(
                context,
                TravelerMinor(age: age, foreign: foreign),
              ),
              child: const Text('Agregar'),
            ),
          ],
        ),
      ),
    );
    if (mounted && minor != null) {
      setState(() {
        minors.add(minor);
      });
    }
  }

  Future<void> submit() async {
    if (busy || confirming || priceChanged || availabilityChanged) return;
    final api = context.read<AndariaApi>();
    final accountId = api.user?['id'];
    if (!api.signedIn || accountId == null) return;
    if (nationals + foreigners < 1 || nationals + foreigners > 100) {
      message(context, const ApiFailure('Selecciona entre 1 y 100 adultos.'));
      return;
    }
    if (price.capacity > widget.departure.n('available_capacity')) {
      message(
        context,
        const ApiFailure('Tu grupo supera los cupos disponibles.'),
      );
      return;
    }
    if (proof == null || method.isEmpty) {
      message(
        context,
        const ApiFailure(
          'Selecciona el medio de pago y adjunta tu comprobante.',
        ),
      );
      return;
    }
    confirming = true;
    bool? accepted;
    try {
      accepted = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Enviar compra a revisión'),
          content: Text(
            'Enviarás un comprobante por ${money(price.total)} para ${price.travelers} viajeros. La agencia revisará el pago antes de confirmarlo.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Revisar'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Enviar'),
            ),
          ],
        ),
      );
    } finally {
      confirming = false;
    }
    if (accepted != true || !mounted) return;
    if (!api.signedIn || api.user?['id'] != accountId) return;
    final payload = {
      'departure_id': widget.departure.id,
      'payment_method': method,
      'national_adults': nationals,
      'foreign_adults': foreigners,
      'minors': minors.map((e) => e.toJson()).toList(),
      'payment_proof': proof!.dataUrl,
      'expected_total_cents': price.total,
    };
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final storageKey = PurchaseKeys.storageKey(
        api.client.options.baseUrl,
        integer(api.user?['id']),
        payload,
      );
      final requestKey = await PurchaseKeys.obtain(storageKey);
      if (!api.signedIn || api.user?['id'] != accountId) {
        throw const ApiFailure(
          'La sesión cambió. Vuelve a revisar la compra.',
          code: 'SESSION_CHANGED',
        );
      }
      final data = await api.request(
        '/me/purchases',
        method: 'POST',
        body: payload,
        authenticated: true,
        idempotencyKey: requestKey,
      );
      // Cleanup failure must never turn an accepted purchase into a failed one.
      try {
        await PurchaseKeys.accepted(storageKey);
      } catch (_) {}
      if (!mounted) return;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (_) => PurchaseDetailPage(id: integer(data['id'])),
        ),
      );
    } catch (e) {
      if (mounted) {
        setState(() {
          priceChanged = e is ApiFailure && e.code == 'PURCHASE_PRICE_CHANGED';
          availabilityChanged =
              e is ApiFailure &&
              [
                'CAPACITY_UNAVAILABLE',
                'DEPARTURE_NOT_BOOKABLE',
              ].contains(e.code);
          error =
              '${e is ApiFailure ? e.message : 'No se pudo enviar la compra.'}${e is ApiFailure && e.status == null ? ' Puedes reintentar sin duplicar esta solicitud.' : ''}';
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Tu experiencia')),
    body: AccountScope(
      child: options == null
          ? (error != null
                ? EmptyView(
                    icon: Icons.wifi_off_outlined,
                    title: 'No pudimos cargar los medios de pago',
                    description: error!,
                    action: 'Volver a intentar',
                    onAction: loadOptions,
                  )
                : const Center(child: CircularProgressIndicator()))
          : AbsorbPointer(
              absorbing: busy,
              child: ListView(
                padding: const EdgeInsets.all(20),
                children: [
                  Text(
                    widget.item.name,
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  Fact(
                    Icons.calendar_today_outlined,
                    dateLabel(widget.departure.s('starts_at')),
                  ),
                  Fact(
                    Icons.location_on_outlined,
                    widget.departure.s('meeting_point'),
                  ),
                  Fact(
                    Icons.access_time,
                    'Encuentro: ${dateLabel(widget.departure.s('meeting_at'))}',
                  ),
                  if (widget.departure.s('instructions').isNotEmpty)
                    Notice(widget.departure.s('instructions')),
                  const BuyerSummary(),
                  const SectionTitle('¿Quiénes viajarán?'),
                  Counter(
                    label: 'Adultos nacionales',
                    value: nationals,
                    max: 100 - foreigners,
                    onChanged: (v) => setState(() {
                      nationals = v;
                    }),
                  ),
                  Counter(
                    label: 'Adultos extranjeros',
                    value: foreigners,
                    max: 100 - nationals,
                    onChanged: (v) => setState(() {
                      foreigners = v;
                    }),
                  ),
                  ...minors.asMap().entries.map(
                    (entry) => ListTile(
                      contentPadding: EdgeInsets.zero,
                      title: Text('Menor de ${entry.value.age} años'),
                      subtitle: Text(
                        entry.value.foreign ? 'Extranjero' : 'Nacional',
                      ),
                      trailing: IconButton(
                        tooltip: 'Quitar menor',
                        icon: const Icon(Icons.close),
                        onPressed: () => setState(() {
                          minors.removeAt(entry.key);
                        }),
                      ),
                    ),
                  ),
                  OutlinedButton.icon(
                    onPressed: minors.length >= 100 ? null : addMinor,
                    icon: const Icon(Icons.add),
                    label: const Text('Agregar menor'),
                  ),
                  Notice(
                    'Los menores de ${options!['minimum_paying_age']} años no pagan ni descuentan cupo.',
                  ),
                  const SectionTitle('Resumen del precio'),
                  Fact(
                    Icons.payments_outlined,
                    'Tarifa nacional: ${money(widget.item.price)}',
                  ),
                  if (widget.item.n('foreign_surcharge_cents') > 0)
                    Fact(
                      Icons.payments_outlined,
                      'Adicional por extranjero que paga: ${money(widget.item.n('foreign_surcharge_cents'))}',
                    ),
                  Text(
                    '${countLabel(price.travelers, 'viajero', 'viajeros')} · ${countLabel(price.capacity, 'cupo', 'cupos')} · ${countLabel(price.freeMinors, 'menor sin pago', 'menores sin pago')}',
                  ),
                  const SizedBox(height: 12),
                  Text(
                    'Total: ${money(price.total)}',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const Notice(
                    'El precio y los cupos se verifican al enviar la compra. No hay una reserva previa durante la transferencia.',
                    warning: true,
                  ),
                  const SectionTitle('Cancelación por el turista'),
                  Text(
                    widget.item.b('cancellation_allowed')
                        ? 'Esta experiencia permite solicitar la cancelación con al menos ${widget.item.n('cancellation_notice_hours')} horas de anticipación a la salida, según su política.'
                        : 'Esta experiencia no permite cancelaciones solicitadas por el turista.',
                  ),
                  const SectionTitle('Realiza el pago'),
                  if (strings(options!['methods']).isEmpty)
                    const Notice(
                      'La agencia no tiene medios de pago disponibles.',
                      warning: true,
                    ),
                  Wrap(
                    spacing: 8,
                    children: strings(options!['methods'])
                        .map(
                          (v) => ChoiceChip(
                            label: Text(v == 'qr' ? 'QR' : 'Transferencia'),
                            selected: method == v,
                            onSelected: (_) => setState(() {
                              method = v;
                            }),
                          ),
                        )
                        .toList(),
                  ),
                  if (method == 'qr' && options!['qr_image'] is String) ...[
                    const SizedBox(height: 16),
                    Image.memory(
                      base64Decode(
                        (options!['qr_image'] as String).split(',').last,
                      ),
                      height: 230,
                      errorBuilder: (_, _, _) =>
                          const Notice('No se pudo mostrar el QR.'),
                    ),
                    OutlinedButton.icon(
                      onPressed: () => shareImage(
                        context,
                        base64Decode(
                          (options!['qr_image'] as String).split(',').last,
                        ),
                        'andaria-pago',
                      ),
                      icon: const Icon(Icons.share_outlined),
                      label: const Text('Compartir QR para pagar'),
                    ),
                  ],
                  if (method == 'transfer') ...[
                    Fact(
                      Icons.account_balance_outlined,
                      options!['bank_name']?.toString() ?? '',
                    ),
                    SelectableText(
                      'Titular: ${options!['account_holder']}\nCuenta: ${options!['account_number']}',
                    ),
                  ],
                  if ((options!['payment_instructions']?.toString() ?? '')
                      .isNotEmpty)
                    Notice(options!['payment_instructions'].toString()),
                  const SectionTitle('Adjunta tu comprobante'),
                  ProofPicker(image: proof, onPick: pick),
                  const Notice(
                    'Enviar el comprobante no confirma el pago. La agencia debe revisarlo.',
                  ),
                  if (error != null) Notice(error!, warning: true),
                  if (priceChanged || availabilityChanged) ...[
                    const Notice(
                      'Si ya transferiste el dinero, consulta a la agencia para resolver el pago antes de realizar otra transferencia.',
                      warning: true,
                    ),
                    OutlinedButton(
                      onPressed: () => Navigator.of(context).pop(),
                      child: Text(
                        priceChanged
                            ? 'Revisar precio actualizado'
                            : 'Revisar salidas disponibles',
                      ),
                    ),
                  ],
                  const SizedBox(height: 24),
                  BusyButton(
                    label: 'Enviar compra · ${money(price.total)}',
                    busy: busy,
                    onPressed:
                        options != null &&
                            method.isNotEmpty &&
                            !priceChanged &&
                            !availabilityChanged
                        ? submit
                        : null,
                  ),
                  AgencyContact(
                    name: options!['agency_name']?.toString() ?? '',
                    phone: options!['agency_phone']?.toString() ?? '',
                    email: options!['agency_email']?.toString() ?? '',
                  ),
                  const SizedBox(height: 24),
                ],
              ),
            ),
    ),
  );
}

class Counter extends StatelessWidget {
  const Counter({
    super.key,
    required this.label,
    required this.value,
    required this.max,
    required this.onChanged,
  });
  final String label;
  final int value, max;
  final ValueChanged<int> onChanged;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 6),
    child: Row(
      children: [
        Expanded(child: Text(label)),
        IconButton(
          tooltip: 'Reducir $label',
          onPressed: value > 0 ? () => onChanged(value - 1) : null,
          icon: const Icon(Icons.remove_circle_outline),
        ),
        SizedBox(width: 32, child: Text('$value', textAlign: TextAlign.center)),
        IconButton(
          tooltip: 'Agregar $label',
          onPressed: value < max ? () => onChanged(value + 1) : null,
          icon: const Icon(Icons.add_circle_outline),
        ),
      ],
    ),
  );
}

class BuyerSummary extends StatelessWidget {
  const BuyerSummary({super.key});
  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    final account = api.user ?? <String, dynamic>{};
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SectionTitle('Datos del comprador'),
        Text(api.fullName, style: Theme.of(context).textTheme.titleMedium),
        Text(account['email']?.toString() ?? ''),
        if ((account['phone']?.toString() ?? '').isNotEmpty)
          Text('Teléfono: ${account['phone']}'),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: () => openPage(context, const EditProfilePage()),
          icon: const Icon(Icons.person_outline),
          label: const Text('Revisar mis datos'),
        ),
        const SizedBox(height: 8),
        const Text(
          'La compra conservará estos datos al enviarla.',
          style: TextStyle(fontSize: 12),
        ),
      ],
    );
  }
}
