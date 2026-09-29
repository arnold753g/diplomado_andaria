import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../core/theme.dart';

class PhotoGalleryPage extends StatefulWidget {
  const PhotoGalleryPage({
    super.key,
    required this.type,
    required this.id,
    required this.photos,
    this.initialIndex = 0,
  });
  final String type;
  final int id, initialIndex;
  final List<int> photos;
  @override
  State<PhotoGalleryPage> createState() => _PhotoGalleryPageState();
}

class _PhotoGalleryPageState extends State<PhotoGalleryPage> {
  late int index = widget.initialIndex;
  int retry = 0;
  @override
  Widget build(BuildContext context) {
    final api = context.read<AndariaApi>();
    return Scaffold(
      backgroundColor: ink,
      appBar: AppBar(
        backgroundColor: ink,
        foregroundColor: Colors.white,
        title: const Text('Fotografías'),
      ),
      body: Column(
        children: [
          Expanded(
            child: LayoutBuilder(
              builder: (context, constraints) => InteractiveViewer(
                key: ValueKey('$index-$retry'),
                minScale: 1,
                maxScale: 5,
                child: SizedBox(
                  width: constraints.maxWidth,
                  height: constraints.maxHeight,
                  child: Image.network(
                    api.photoUrl(widget.type, widget.id, widget.photos[index]),
                    fit: BoxFit.contain,
                    semanticLabel:
                        'Fotografía ${index + 1} de ${widget.photos.length}',
                    loadingBuilder: (_, child, progress) => progress == null
                        ? child
                        : const Center(
                            child: CircularProgressIndicator(color: lime),
                          ),
                    errorBuilder: (_, _, _) => Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Text(
                            'No se pudo cargar la fotografía.',
                            style: TextStyle(color: Colors.white),
                          ),
                          TextButton(
                            onPressed: () => setState(() {
                              retry++;
                            }),
                            child: const Text(
                              'Volver a intentar',
                              style: TextStyle(color: lime),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
          SafeArea(
            top: false,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      IconButton(
                        tooltip: 'Fotografía anterior',
                        color: Colors.white,
                        disabledColor: muted,
                        onPressed: index > 0
                            ? () => setState(() {
                                index--;
                              })
                            : null,
                        icon: const Icon(Icons.chevron_left),
                      ),
                      Semantics(
                        liveRegion: true,
                        child: Text(
                          '${index + 1} de ${widget.photos.length}',
                          style: const TextStyle(color: Colors.white),
                        ),
                      ),
                      IconButton(
                        tooltip: 'Fotografía siguiente',
                        color: Colors.white,
                        disabledColor: muted,
                        onPressed: index + 1 < widget.photos.length
                            ? () => setState(() {
                                index++;
                              })
                            : null,
                        icon: const Icon(Icons.chevron_right),
                      ),
                    ],
                  ),
                  const Text(
                    'Pellizca la fotografía para ampliar',
                    style: TextStyle(color: Color(0xFFBAC1C5), fontSize: 12),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
