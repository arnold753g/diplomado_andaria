import 'account_scope.dart';
import 'dart:convert';
import 'dart:typed_data';
import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:share_plus/share_plus.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import 'widgets.dart';

class ProofImage {
  const ProofImage(this.bytes, this.mime);
  final Uint8List bytes;
  final String mime;
  String get dataUrl => 'data:$mime;base64,${base64Encode(bytes)}';
  static Future<ProofImage?> pick() async {
    final file = await ImagePicker().pickImage(source: ImageSource.gallery);
    if (file == null) return null;
    final bytes = await file.readAsBytes();
    return validate(bytes);
  }

  static Future<ProofImage> validate(Uint8List bytes) async {
    final proof = fromBytes(bytes);
    final buffer = await ui.ImmutableBuffer.fromUint8List(bytes);
    ui.ImageDescriptor? descriptor;
    ui.Codec? codec;
    try {
      descriptor = await ui.ImageDescriptor.encoded(buffer);
      if (descriptor.width > 4096 || descriptor.height > 4096) {
        throw const ApiFailure(
          'Selecciona una imagen de hasta 4096 × 4096 píxeles.',
        );
      }
      codec = await descriptor.instantiateCodec();
      final frame = await codec.getNextFrame();
      frame.image.dispose();
      return proof;
    } on ApiFailure {
      rethrow;
    } catch (_) {
      throw const ApiFailure(
        'No se pudo leer la imagen. Selecciona otro PNG o JPG.',
      );
    } finally {
      codec?.dispose();
      descriptor?.dispose();
      buffer.dispose();
    }
  }

  static ProofImage fromBytes(Uint8List bytes) {
    if (bytes.length > 5 * 1024 * 1024) {
      throw const ApiFailure('Selecciona una imagen de hasta 5 MB.');
    }
    final png =
        bytes.length >= 8 &&
        bytes[0] == 137 &&
        bytes[1] == 80 &&
        bytes[2] == 78 &&
        bytes[3] == 71 &&
        bytes[4] == 13 &&
        bytes[5] == 10 &&
        bytes[6] == 26 &&
        bytes[7] == 10;
    final jpg =
        bytes.length >= 3 &&
        bytes[0] == 255 &&
        bytes[1] == 216 &&
        bytes[2] == 255;
    if (!png && !jpg) {
      throw const ApiFailure('Selecciona una imagen PNG o JPG.');
    }
    return ProofImage(bytes, png ? 'image/png' : 'image/jpeg');
  }
}

Future<void> shareImage(
  BuildContext context,
  Uint8List bytes,
  String name,
) async {
  try {
    await SharePlus.instance.share(
      ShareParams(
        files: [
          XFile.fromData(bytes, mimeType: 'image/png', name: '$name.png'),
        ],
        fileNameOverrides: ['$name.png'],
      ),
    );
  } catch (e) {
    if (context.mounted) message(context, e);
  }
}

class ImageViewerPage extends StatelessWidget {
  const ImageViewerPage({super.key, required this.bytes, required this.title});
  final Uint8List bytes;
  final String title;
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(title)),
    body: AccountScope(
      child: !context.watch<AndariaApi>().signedIn
          ? const EmptyView(
              icon: Icons.lock_outline,
              title: 'La sesión terminó',
              description: 'Vuelve a iniciar sesión para ver el comprobante.',
            )
          : Center(
              child: InteractiveViewer(
                minScale: .5,
                maxScale: 5,
                child: Image.memory(
                  bytes,
                  errorBuilder: (_, _, _) =>
                      const Text('La imagen no se pudo mostrar.'),
                ),
              ),
            ),
    ),
  );
}

class ProofPicker extends StatelessWidget {
  const ProofPicker({
    super.key,
    required this.image,
    required this.onPick,
    this.label = 'Seleccionar comprobante',
    this.enabled = true,
  });
  final ProofImage? image;
  final VoidCallback onPick;
  final String label;
  final bool enabled;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      if (image != null)
        Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.memory(
              image!.bytes,
              height: 180,
              cacheHeight: 640,
              width: double.infinity,
              fit: BoxFit.contain,
            ),
          ),
        ),
      if (image != null)
        TextButton.icon(
          onPressed: enabled
              ? () => openPage(
                  context,
                  ImageViewerPage(
                    bytes: image!.bytes,
                    title: 'Revisa tu comprobante',
                  ),
                )
              : null,
          icon: const Icon(Icons.zoom_in),
          label: const Text('Revisar imagen'),
        ),
      SizedBox(
        width: double.infinity,
        child: OutlinedButton.icon(
          onPressed: enabled ? onPick : null,
          icon: const Icon(Icons.image_outlined),
          label: Text(image == null ? label : 'Cambiar imagen'),
        ),
      ),
      const SizedBox(height: 8),
      const Text(
        'PNG o JPG · máximo 5 MB y 4096 × 4096 píxeles',
        style: TextStyle(fontSize: 12),
      ),
    ],
  );
}
