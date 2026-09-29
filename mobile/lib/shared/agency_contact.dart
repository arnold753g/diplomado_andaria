import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../core/api.dart';
import 'widgets.dart';

class AgencyContact extends StatelessWidget {
  const AgencyContact({
    super.key,
    required this.name,
    required this.phone,
    required this.email,
    this.reference = '',
  });
  final String name, phone, email, reference;

  Future<void> open(BuildContext context, Uri uri) async {
    try {
      if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
        throw const ApiFailure(
          'No se encontró una aplicación para abrir este contacto. Puedes copiar el dato de la agencia.',
        );
      }
    } catch (error) {
      if (context.mounted) message(context, error);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (phone.trim().isEmpty && email.trim().isEmpty) {
      return const SizedBox.shrink();
    }
    final dial = phone.replaceAll(RegExp(r'[^0-9+]'), '');
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SectionTitle('Contacta a la agencia'),
        Text(name, style: Theme.of(context).textTheme.titleMedium),
        if (reference.isNotEmpty)
          SelectableText('Referencia de compra: $reference'),
        const SizedBox(height: 8),
        if (phone.trim().isNotEmpty) ...[
          SelectableText(phone),
          if (RegExp(r'^[+0-9\s().-]+$').hasMatch(phone) &&
              RegExp(r'^\+?[0-9]{5,15}$').hasMatch(dial))
            OutlinedButton.icon(
              onPressed: () => open(context, Uri(scheme: 'tel', path: dial)),
              icon: const Icon(Icons.phone_outlined),
              label: const Text('Abrir teléfono'),
            ),
        ],
        if (email.trim().isNotEmpty) ...[
          SelectableText(email),
          if (RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$').hasMatch(email.trim()))
            OutlinedButton.icon(
              onPressed: () => open(
                context,
                Uri(
                  scheme: 'mailto',
                  path: email.trim(),
                  query:
                      'subject=${Uri.encodeComponent(reference.isEmpty ? 'Consulta Andaria' : 'Consulta sobre la compra $reference')}',
                ),
              ),
              icon: const Icon(Icons.mail_outline),
              label: const Text('Redactar correo'),
            ),
        ],
        const Text('La agencia revisa los pagos y gestiona las devoluciones.'),
      ],
    );
  }
}
