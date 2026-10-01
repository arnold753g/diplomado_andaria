import '../shared/account_scope.dart';
import '../shared/agency_contact.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../core/models.dart';
import '../core/theme.dart';
import '../shared/widgets.dart';
import '../shared/images.dart';
import 'auth.dart';

class PurchasesPage extends StatefulWidget {
  const PurchasesPage({super.key});
  @override
  State<PurchasesPage> createState() => _PurchasesPageState();
}

class _PurchasesPageState extends State<PurchasesPage> {
  int page = 1;
  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    return Scaffold(
      appBar: AppBar(title: const Text('Mis compras')),
      body: !api.signedIn
          ? EmptyView(
              icon: Icons.confirmation_number_outlined,
              title: 'Tus próximas experiencias',
              description:
                  'Inicia sesión para consultar compras, pagos y devoluciones.',
              action: 'Iniciar sesión',
              onAction: () => requireTourist(context),
            )
          : AsyncContent<Json>(
              key: ValueKey('$page-${api.user?['id']}'),
              load: () async => Map<String, dynamic>.from(
                await api.request(
                      '/me/purchases',
                      query: {'page': page, 'limit': 20},
                      authenticated: true,
                    )
                    as Map,
              ),
              builder: (data, reload) {
                final purchases = objects(
                  data['purchases'],
                ).map(Purchase.new).toList();
                if (purchases.isEmpty) {
                  return EmptyView(
                    action: page > 1 ? 'Volver a la primera página' : null,
                    onAction: page > 1
                        ? () => setState(() {
                            page = 1;
                          })
                        : null,
                    icon: Icons.confirmation_number_outlined,
                    title: 'Todavía no tienes compras',
                    description:
                        'Explora una experiencia y elige una salida disponible.',
                  );
                }
                return RefreshIndicator(
                  onRefresh: reload,
                  child: ListView(
                    padding: const EdgeInsets.all(20),
                    children: [
                      ...purchases.map(
                        (item) => Padding(
                          padding: const EdgeInsets.only(bottom: 16),
                          child: Material(
                            color: canvas,
                            borderRadius: BorderRadius.circular(18),
                            clipBehavior: Clip.antiAlias,
                            child: InkWell(
                              onTap: () async {
                                await openPage(
                                  context,
                                  PurchaseDetailPage(id: item.id),
                                );
                                await reload();
                              },
                              child: Padding(
                                padding: const EdgeInsets.all(18),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      item.s('reference'),
                                      style: const TextStyle(
                                        color: muted,
                                        fontSize: 12,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Text(
                                      item.s('package_name'),
                                      style: Theme.of(
                                        context,
                                      ).textTheme.titleLarge,
                                    ),
                                    const SizedBox(height: 8),
                                    StatusBadge(item),
                                    Fact(
                                      Icons.calendar_today_outlined,
                                      dateLabel(item.s('departure_start')),
                                    ),
                                    Text(
                                      money(item.n('total_cents')),
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    const Text('Ver compra →'),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          TextButton(
                            onPressed: page > 1
                                ? () => setState(() {
                                    page--;
                                  })
                                : null,
                            child: const Text('Anterior'),
                          ),
                          Text('Página $page'),
                          TextButton(
                            onPressed:
                                page * 20 < integer(data['pagination']['total'])
                                ? () => setState(() {
                                    page++;
                                  })
                                : null,
                            child: const Text('Siguiente'),
                          ),
                        ],
                      ),
                    ],
                  ),
                );
              },
            ),
    );
  }
}

class StatusBadge extends StatelessWidget {
  const StatusBadge(this.purchase, {super.key});
  final Purchase purchase;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
    decoration: BoxDecoration(
      color: ['confirmed', 'refunded'].contains(purchase.status)
          ? const Color(0xFFEAF7D2)
          : const Color(0xFFFFF0D0),
      borderRadius: BorderRadius.circular(8),
    ),
    child: Text(
      purchase.statusLabel,
      style: const TextStyle(fontWeight: FontWeight.w500),
    ),
  );
}

class PurchaseDetailPage extends StatelessWidget {
  const PurchaseDetailPage({super.key, required this.id});
  final int id;
  Future<void> viewImage(
    BuildContext context,
    String suffix,
    String title,
  ) async {
    try {
      final bytes = await context.read<AndariaApi>().privateImage(
        '/me/purchases/$id/$suffix',
      );
      if (context.mounted) {
        await openPage(context, ImageViewerPage(bytes: bytes, title: title));
      }
    } catch (e) {
      if (context.mounted) message(context, e);
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    return Scaffold(
      appBar: AppBar(title: const Text('Detalle de compra')),
      body: !api.signedIn
          ? EmptyView(
              icon: Icons.lock_outline,
              title: 'Inicia sesión nuevamente',
              description:
                  api.sessionNotice ??
                  'Necesitas una sesión activa para consultar esta compra.',
              action: 'Iniciar sesión',
              onAction: () => requireTourist(context),
            )
          : AsyncContent<Purchase>(
              key: ValueKey(api.user?['id']),
              load: () async => Purchase(
                Map<String, dynamic>.from(
                  await api.request('/me/purchases/$id', authenticated: true)
                      as Map,
                ),
              ),
              builder: (item, reload) => RefreshIndicator(
                onRefresh: reload,
                child: ListView(
                  padding: const EdgeInsets.all(20),
                  children: [
                    Text(
                      item.s('reference'),
                      style: const TextStyle(color: muted),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      item.s('package_name'),
                      style: Theme.of(context).textTheme.headlineMedium,
                    ),
                    const SizedBox(height: 16),
                    Align(
                      alignment: Alignment.centerLeft,
                      child: StatusBadge(item),
                    ),
                    const SizedBox(height: 16),
                    Fact(Icons.storefront_outlined, item.s('agency_name')),
                    Fact(
                      Icons.calendar_today_outlined,
                      dateLabel(item.s('departure_start')),
                    ),
                    Fact(Icons.location_on_outlined, item.s('meeting_point')),
                    Fact(
                      Icons.people_outline,
                      '${countLabel(item.n('capacity_count'), 'cupo', 'cupos')} · ${countLabel(item.n('free_minor_count'), 'menor sin pago', 'menores sin pago')}',
                    ),
                    Fact(
                      Icons.payments_outlined,
                      '${money(item.n('total_cents'))} · ${item.s('payment_method') == 'qr' ? 'QR' : 'Transferencia'}',
                    ),
                    if (item.status == 'payment_review')
                      const Notice(
                        'La agencia está revisando el pago. Tus cupos se mantienen mientras se revisa el comprobante.',
                      ),
                    if (item.s('rejection_reason').isNotEmpty)
                      Notice(item.s('rejection_reason'), warning: true),
                    if (item.b('has_proof'))
                      OutlinedButton.icon(
                        onPressed: () =>
                            viewImage(context, 'proof', 'Comprobante de pago'),
                        icon: const Icon(Icons.receipt_long_outlined),
                        label: const Text('Ver comprobante de pago'),
                      ),
                    if (item.status == 'correction_requested') ...[
                      const Notice(
                        'Tus cupos se mantienen mientras corriges el comprobante.',
                      ),
                      FilledButton(
                        onPressed: () async {
                          await openPage(
                            context,
                            CorrectProofPage(purchase: item),
                          );
                          await reload();
                        },
                        child: const Text('Corregir comprobante'),
                      ),
                    ],
                    if (item.b('cancellable')) ...[
                      const SectionTitle('Cancelación'),
                      Text(
                        'Puedes cancelar hasta ${dateLabel(item.s('cancellation_deadline'))} con devolución del 100 %.',
                      ),
                      const SizedBox(height: 12),
                      OutlinedButton(
                        onPressed: () async {
                          await openPage(
                            context,
                            RefundPage(purchase: item, cancellation: true),
                          );
                          await reload();
                        },
                        child: const Text('Cancelar compra'),
                      ),
                    ],
                    if (item.s('cancellation_reason').isNotEmpty)
                      Notice(
                        'Motivo de cancelación: ${item.s('cancellation_reason')}',
                      ),
                    if ([
                      'refund_pending',
                      'refunded',
                    ].contains(item.status)) ...[
                      const SectionTitle('Tu devolución'),
                      if (item.s('refund_reason').isNotEmpty)
                        Text(item.s('refund_reason')),
                      if (item.status == 'refund_pending')
                        Notice(
                          '${item.b('refund_overdue') ? 'La devolución está atrasada. ' : ''}La agencia debe registrar el reembolso hasta ${dateLabel(item.s('refund_due_at'))}.',
                          warning: true,
                        ),
                      if (item.s('refund_method') == 'bank_transfer')
                        SelectableText(
                          'Banco: ${item.s('refund_bank_name')}\nTitular: ${item.s('refund_account_holder')}\nCuenta: ${item.s('refund_account_number')}',
                        ),
                      if (item.b('has_refund_qr'))
                        OutlinedButton(
                          onPressed: () => viewImage(
                            context,
                            'refund-qr',
                            'QR para devolución',
                          ),
                          child: const Text('Ver QR para devolución'),
                        ),
                      if (item.status == 'refund_pending')
                        FilledButton(
                          onPressed: () async {
                            await openPage(context, RefundPage(purchase: item));
                            await reload();
                          },
                          child: const Text('Actualizar datos de devolución'),
                        ),
                      if (item.status == 'refunded') ...[
                        Notice(
                          'Devolución registrada el ${dateLabel(item.s('refunded_at'))}. ${item.s('refund_reference')}',
                        ),
                        if (item.b('has_refund_proof'))
                          OutlinedButton(
                            onPressed: () => viewImage(
                              context,
                              'refund-proof',
                              'Comprobante de devolución',
                            ),
                            child: const Text('Ver comprobante de devolución'),
                          ),
                      ],
                    ],
                    const SizedBox(height: 24),
                    const Text(
                      'Horarios de Bolivia (UTC−4).',
                      style: TextStyle(color: muted),
                    ),
                    AgencyContact(
                      name: item.s('agency_name'),
                      phone: item.s('agency_phone'),
                      email: item.s('agency_email'),
                      reference: item.s('reference'),
                    ),
                  ],
                ),
              ),
            ),
    );
  }
}

class CorrectProofPage extends StatefulWidget {
  const CorrectProofPage({super.key, required this.purchase});
  final Purchase purchase;
  @override
  State<CorrectProofPage> createState() => _CorrectProofPageState();
}

class _CorrectProofPageState extends State<CorrectProofPage> {
  ProofImage? image;
  bool busy = false;
  String? error;
  Future<void> save() async {
    if (image == null) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      await context.read<AndariaApi>().request(
        '/me/purchases/${widget.purchase.id}/proof',
        method: 'PATCH',
        authenticated: true,
        body: {
          'version': widget.purchase.n('version'),
          'payment_proof': image!.dataUrl,
        },
      );
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e.toString();
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
    appBar: AppBar(title: const Text('Corregir comprobante')),
    body: AccountScope(
      child: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Notice(widget.purchase.s('rejection_reason'), warning: true),
          ProofPicker(
            image: image,
            enabled: !busy,
            onPick: () async {
              try {
                final picked = await ProofImage.pick();
                if (mounted && picked != null) {
                  setState(() {
                    image = picked;
                  });
                }
              } catch (e) {
                if (context.mounted) message(context, e);
              }
            },
          ),
          if (error != null) Notice(error!, warning: true),
          const SizedBox(height: 24),
          BusyButton(
            label: 'Enviar corrección',
            busy: busy,
            onPressed: image == null ? null : save,
          ),
        ],
      ),
    ),
  );
}

class RefundPage extends StatefulWidget {
  const RefundPage({
    super.key,
    required this.purchase,
    this.cancellation = false,
  });
  final Purchase purchase;
  final bool cancellation;
  @override
  State<RefundPage> createState() => _RefundPageState();
}

class _RefundPageState extends State<RefundPage> {
  final form = GlobalKey<FormState>();
  final reason = TextEditingController(),
      bank = TextEditingController(),
      holder = TextEditingController(),
      account = TextEditingController();
  String method = 'bank_transfer';
  ProofImage? image;
  bool busy = false;
  String? error;
  @override
  void initState() {
    super.initState();
    bank.text = widget.purchase.s('refund_bank_name');
    holder.text = widget.purchase.s('refund_account_holder');
    account.text = widget.purchase.s('refund_account_number');
  }

  @override
  void dispose() {
    reason.dispose();
    bank.dispose();
    holder.dispose();
    account.dispose();
    super.dispose();
  }

  Future<void> save() async {
    if (!form.currentState!.validate()) return;
    if (method == 'qr' && image == null) {
      message(
        context,
        const ApiFailure('Selecciona un QR para recibir tu devolución.'),
      );
      return;
    }
    if (widget.cancellation) {
      final accepted = await showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Confirmar cancelación'),
          content: const Text(
            'Se liberarán tus cupos inmediatamente. La agencia tendrá 72 horas para registrar la devolución del pago.',
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Volver'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Confirmar cancelación'),
            ),
          ],
        ),
      );
      if (accepted != true || !mounted) return;
    }
    setState(() {
      busy = true;
      error = null;
    });
    try {
      await context.read<AndariaApi>().request(
        '/me/purchases/${widget.purchase.id}/${widget.cancellation ? 'cancel' : 'refund-destination'}',
        method: widget.cancellation ? 'POST' : 'PATCH',
        authenticated: true,
        body: {
          'version': widget.purchase.n('version'),
          if (widget.cancellation) 'reason': reason.text.trim(),
          'refund_method': method,
          if (method == 'qr') 'refund_qr': image!.dataUrl,
          if (method == 'bank_transfer') ...{
            'refund_bank_name': bank.text.trim(),
            'refund_account_holder': holder.text.trim(),
            'refund_account_number': account.text.trim(),
          },
        },
      );
      if (mounted) Navigator.pop(context);
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e.toString();
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
    appBar: AppBar(
      title: Text(
        widget.cancellation ? 'Cancelar compra' : 'Datos para devolución',
      ),
    ),
    body: AccountScope(
      child: AbsorbPointer(
        absorbing: busy,
        child: Form(
          key: form,
          child: ListView(
            padding: const EdgeInsets.all(20),
            children: [
              if (widget.cancellation) ...[
                const Notice(
                  'Al confirmar se liberarán los cupos y comenzará el plazo de devolución de 72 horas.',
                  warning: true,
                ),
                TextFormField(
                  controller: reason,
                  maxLines: 3,
                  maxLength: 1000,
                  decoration: const InputDecoration(
                    labelText: 'Motivo de cancelación',
                  ),
                  validator: (v) => (v?.trim().length ?? 0) < 5
                      ? 'Escribe al menos 5 caracteres.'
                      : null,
                ),
              ],
              const SectionTitle('¿Dónde recibirás tu devolución?'),
              Wrap(
                spacing: 8,
                children: [
                  ChoiceChip(
                    label: const Text('Cuenta bancaria'),
                    selected: method == 'bank_transfer',
                    onSelected: (_) => setState(() {
                      method = 'bank_transfer';
                    }),
                  ),
                  ChoiceChip(
                    label: const Text('QR'),
                    selected: method == 'qr',
                    onSelected: (_) => setState(() {
                      method = 'qr';
                    }),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              if (method == 'qr')
                ProofPicker(
                  image: image,
                  label: 'Seleccionar QR',
                  onPick: () async {
                    try {
                      final picked = await ProofImage.pick();
                      if (mounted && picked != null) {
                        setState(() {
                          image = picked;
                        });
                      }
                    } catch (e) {
                      if (context.mounted) message(context, e);
                    }
                  },
                )
              else ...[
                TextFormField(
                  controller: bank,
                  maxLength: 120,
                  decoration: const InputDecoration(labelText: 'Banco'),
                  validator: (v) =>
                      (v?.trim().isEmpty ?? true) ? 'Ingresa el banco.' : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: holder,
                  maxLength: 160,
                  decoration: const InputDecoration(
                    labelText: 'Titular de la cuenta',
                  ),
                  validator: (v) => (v?.trim().isEmpty ?? true)
                      ? 'Ingresa el titular.'
                      : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: account,
                  maxLength: 80,
                  decoration: const InputDecoration(
                    labelText: 'Número de cuenta',
                  ),
                  validator: (v) => (v?.trim().isEmpty ?? true)
                      ? 'Ingresa el número de cuenta.'
                      : null,
                ),
              ],
              if (error != null) Notice(error!, warning: true),
              const SizedBox(height: 24),
              BusyButton(
                label: widget.cancellation
                    ? 'Confirmar cancelación'
                    : 'Guardar datos de devolución',
                busy: busy,
                onPressed: save,
              ),
            ],
          ),
        ),
      ),
    ),
  );
}
