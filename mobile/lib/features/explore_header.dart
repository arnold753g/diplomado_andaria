import 'package:flutter/material.dart';
import '../core/theme.dart';
import '../shared/brand.dart';

class ExploreHeader extends StatelessWidget {
  const ExploreHeader({super.key, required this.search});
  final Widget search;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      const AndariaBrand(),
      const SizedBox(height: 24),
      Container(
        width: double.infinity,
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: const Color(0xFFF1F5E9),
          borderRadius: BorderRadius.circular(24),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Sal de la rutina.\nDescubre Bolivia.',
              style: Theme.of(context).textTheme.headlineLarge,
            ),
            const SizedBox(height: 12),
            const Text(
              'Tu próxima experiencia, comenzando por Tarija.',
              style: TextStyle(color: muted),
            ),
          ],
        ),
      ),
      const SizedBox(height: 20),
      search,
    ],
  );
}

class CatalogModules extends StatelessWidget {
  const CatalogModules({
    super.key,
    required this.attractions,
    required this.onChanged,
  });
  final bool attractions;
  final ValueChanged<bool> onChanged;

  Widget module(
    BuildContext context, {
    required bool value,
    required IconData icon,
    required String label,
  }) {
    final selected = attractions == value;
    return Expanded(
      child: Semantics(
        button: true,
        selected: selected,
        child: Material(
          color: selected ? lime : canvas,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: BorderSide(color: selected ? lime : line),
          ),
          clipBehavior: Clip.antiAlias,
          child: InkWell(
            onTap: () {
              if (!selected) onChanged(value);
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 16),
              child: Column(
                children: [
                  Icon(icon, size: 24),
                  const SizedBox(height: 8),
                  Text(
                    label,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) => IntrinsicHeight(
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        module(
          context,
          value: false,
          icon: Icons.route_outlined,
          label: 'Paquetes turísticos',
        ),
        const SizedBox(width: 12),
        module(
          context,
          value: true,
          icon: Icons.landscape_outlined,
          label: 'Atracciones turísticas',
        ),
      ],
    ),
  );
}

class CatalogFilterLabel extends StatelessWidget {
  const CatalogFilterLabel({
    super.key,
    required this.icon,
    required this.label,
  });
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) => Container(
    constraints: const BoxConstraints(maxWidth: 320, minHeight: 48),
    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
    decoration: BoxDecoration(
      border: Border.all(color: line),
      borderRadius: BorderRadius.circular(14),
    ),
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 20),
        const SizedBox(width: 8),
        Flexible(child: Text(label)),
        const SizedBox(width: 8),
        const Icon(Icons.expand_more, size: 18),
      ],
    ),
  );
}
