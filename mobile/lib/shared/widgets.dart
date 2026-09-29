import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../core/theme.dart';
import 'photo_gallery.dart';

void message(BuildContext context, Object error) {
  final text = error is ApiFailure
      ? error.message
      : 'No se pudo completar la operación. Inténtalo nuevamente.';
  ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(text)));
}

Future<T?> openPage<T>(BuildContext context, Widget page) =>
    Navigator.of(context).push<T>(MaterialPageRoute(builder: (_) => page));

class AsyncContent<T> extends StatefulWidget {
  const AsyncContent({super.key, required this.load, required this.builder});
  final Future<T> Function() load;
  final Widget Function(T value, Future<void> Function() reload) builder;
  @override
  State<AsyncContent<T>> createState() => _AsyncContentState<T>();
}

class _AsyncContentState<T> extends State<AsyncContent<T>> {
  late Future<T> future;
  @override
  void initState() {
    super.initState();
    future = widget.load();
  }

  Future<void> reload() async {
    if (!mounted) return;
    setState(() {
      future = widget.load();
    });
    try {
      await future;
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) => FutureBuilder<T>(
    future: future,
    builder: (context, snapshot) {
      if (snapshot.connectionState != ConnectionState.done) {
        return const Center(
          child: Padding(
            padding: EdgeInsets.all(48),
            child: CircularProgressIndicator(),
          ),
        );
      }
      if (snapshot.hasError) {
        return EmptyView(
          icon: Icons.wifi_off_rounded,
          title: 'No pudimos cargar el contenido',
          description: snapshot.error is ApiFailure
              ? (snapshot.error as ApiFailure).message
              : 'Comprueba tu conexión e inténtalo nuevamente.',
          action: 'Volver a intentar',
          onAction: reload,
        );
      }
      return widget.builder(snapshot.data as T, reload);
    },
  );
}

class EmptyView extends StatefulWidget {
  const EmptyView({
    super.key,
    required this.icon,
    required this.title,
    required this.description,
    this.action,
    this.onAction,
  });
  final IconData icon;
  final String title, description;
  final String? action;
  final FutureOr<void> Function()? onAction;
  @override
  State<EmptyView> createState() => _EmptyViewState();
}

class _EmptyViewState extends State<EmptyView> {
  bool busy = false;
  Future<void> _runAction() async {
    if (busy || widget.onAction == null) return;
    setState(() {
      busy = true;
    });
    try {
      await widget.onAction!();
    } catch (e) {
      if (mounted) message(context, e);
    } finally {
      if (mounted) {
        setState(() {
          busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => Center(
    child: SingleChildScrollView(
      padding: const EdgeInsets.all(28),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: const BoxDecoration(
              color: canvas,
              shape: BoxShape.circle,
            ),
            child: Icon(widget.icon, size: 40, color: ink),
          ),
          const SizedBox(height: 24),
          Text(
            widget.title,
            style: Theme.of(context).textTheme.titleLarge,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 12),
          Text(
            widget.description,
            textAlign: TextAlign.center,
            style: const TextStyle(color: muted),
          ),
          if (widget.action != null) ...[
            const SizedBox(height: 24),
            BusyButton(
              label: widget.action!,
              busy: busy,
              onPressed: widget.onAction == null ? null : _runAction,
            ),
          ],
        ],
      ),
    ),
  );
}

class Notice extends StatelessWidget {
  const Notice(this.text, {super.key, this.warning = false});
  final String text;
  final bool warning;
  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    margin: const EdgeInsets.symmetric(vertical: 8),
    padding: const EdgeInsets.all(16),
    decoration: BoxDecoration(
      color: warning ? const Color(0xFFFFF5DD) : canvas,
      borderRadius: BorderRadius.circular(14),
    ),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(
          warning ? Icons.info_outline : Icons.check_circle_outline,
          size: 20,
        ),
        const SizedBox(width: 10),
        Expanded(child: Text(text)),
      ],
    ),
  );
}

class SectionTitle extends StatelessWidget {
  const SectionTitle(this.title, {super.key});
  final String title;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 24, bottom: 12),
    child: Text(title, style: Theme.of(context).textTheme.titleLarge),
  );
}

class Fact extends StatelessWidget {
  const Fact(this.icon, this.text, {super.key});
  final IconData icon;
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 6),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 20, color: muted),
        const SizedBox(width: 10),
        Expanded(child: Text(text)),
      ],
    ),
  );
}

class BusyButton extends StatelessWidget {
  const BusyButton({
    super.key,
    required this.label,
    required this.busy,
    required this.onPressed,
  });
  final String label;
  final bool busy;
  final VoidCallback? onPressed;
  @override
  Widget build(BuildContext context) => SizedBox(
    width: double.infinity,
    child: FilledButton(
      onPressed: busy ? null : onPressed,
      child: busy
          ? const SizedBox(
              width: 22,
              height: 22,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : Text(label),
    ),
  );
}

class TravelPhoto extends StatelessWidget {
  const TravelPhoto({
    super.key,
    required this.type,
    required this.id,
    required this.photos,
    this.height = 200,
    this.allowZoom = false,
  });
  final String type;
  final int id;
  final List<int> photos;
  final double height;
  final bool allowZoom;
  @override
  Widget build(BuildContext context) {
    final api = context.read<AndariaApi>();
    Widget placeholder() => Container(
      color: canvas,
      child: const Center(
        child: Icon(Icons.landscape_outlined, color: muted, size: 48),
      ),
    );
    if (photos.isEmpty) return SizedBox(height: height, child: placeholder());
    return SizedBox(
      height: height,
      child: Stack(
        children: [
          PageView(
            children: photos
                .map(
                  (photo) => GestureDetector(
                    onTap: allowZoom
                        ? () => openPage(
                            context,
                            PhotoGalleryPage(
                              type: type,
                              id: id,
                              photos: photos,
                              initialIndex: photos.indexOf(photo),
                            ),
                          )
                        : null,
                    child: Semantics(
                      label: 'Fotografía del destino',
                      image: true,
                      child: Image.network(
                        api.photoUrl(type, id, photo),
                        width: double.infinity,
                        height: height,
                        cacheWidth:
                            (MediaQuery.sizeOf(context).width *
                                    MediaQuery.devicePixelRatioOf(context))
                                .clamp(240, 1600)
                                .round(),
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) => placeholder(),
                        loadingBuilder: (_, child, progress) =>
                            progress == null ? child : placeholder(),
                      ),
                    ),
                  ),
                )
                .toList(),
          ),
          if (photos.length > 1 || allowZoom)
            Positioned(
              right: 12,
              bottom: 12,
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 5,
                ),
                decoration: BoxDecoration(
                  color: ink.withValues(alpha: .75),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  allowZoom
                      ? '${photos.length} ${photos.length == 1 ? 'foto' : 'fotos'} · toca para ampliar'
                      : '${photos.length} fotos · desliza',
                  style: const TextStyle(color: Colors.white, fontSize: 12),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
