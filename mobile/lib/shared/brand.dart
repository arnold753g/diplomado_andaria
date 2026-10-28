import 'package:flutter/material.dart';
import '../core/theme.dart';

class AndariaLogo extends StatelessWidget {
  const AndariaLogo({super.key, this.size = 48});
  final double size;

  @override
  Widget build(BuildContext context) => Image.asset(
    'assets/branding/andaria-symbol.png',
    width: size,
    height: size,
    fit: BoxFit.contain,
    semanticLabel: 'Andaria',
  );
}

class AndariaBrand extends StatelessWidget {
  const AndariaBrand({super.key});

  @override
  Widget build(BuildContext context) => const Row(
    children: [
      ExcludeSemantics(child: AndariaLogo()),
      SizedBox(width: 12),
      Expanded(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Andaria',
              style: TextStyle(
                fontSize: 26,
                fontWeight: FontWeight.w700,
                letterSpacing: -.8,
              ),
            ),
            Text(
              'Turismo en Bolivia',
              style: TextStyle(color: muted, fontSize: 12),
            ),
          ],
        ),
      ),
    ],
  );
}

/// Upright, transparent brand artwork. Motion never delays the real request.
class AndariaLoading extends StatefulWidget {
  const AndariaLoading({
    super.key,
    this.label = 'Cargando contenido',
    this.compact = false,
  });
  final String label;
  final bool compact;

  @override
  State<AndariaLoading> createState() => _AndariaLoadingState();
}

class _AndariaLoadingState extends State<AndariaLoading>
    with SingleTickerProviderStateMixin {
  late final AnimationController controller = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 1400),
  );
  late final CurvedAnimation pulse = CurvedAnimation(
    parent: controller,
    curve: Curves.easeInOut,
  );

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (MediaQuery.disableAnimationsOf(context) ||
        !TickerMode.valuesOf(context).enabled) {
      controller.stop();
      controller.value = .5;
    } else if (!controller.isAnimating) {
      controller.repeat(reverse: true);
    }
  }

  @override
  void dispose() {
    pulse.dispose();
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Semantics(
    label: widget.label,
    liveRegion: true,
    child: ExcludeSemantics(
      child: Padding(
        padding: EdgeInsets.symmetric(
          vertical: widget.compact ? 28 : 48,
          horizontal: 24,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AnimatedBuilder(
              animation: pulse,
              child: AndariaLogo(size: widget.compact ? 56 : 88),
              builder: (_, child) => Opacity(
                opacity: .65 + .35 * pulse.value,
                child: Transform.scale(
                  scale: .94 + .06 * pulse.value,
                  child: child,
                ),
              ),
            ),
            SizedBox(height: widget.compact ? 16 : 24),
            if (!widget.compact) ...[
              const Text(
                'Andaria',
                style: TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w700,
                  letterSpacing: -.8,
                ),
              ),
              const SizedBox(height: 8),
            ],
            Text(
              widget.label,
              textAlign: TextAlign.center,
              style: const TextStyle(color: muted),
            ),
          ],
        ),
      ),
    ),
  );
}
