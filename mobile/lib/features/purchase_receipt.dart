import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:printing/printing.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../shared/account_scope.dart';
import '../shared/widgets.dart';

class PurchaseReceiptPage extends StatelessWidget {
  const PurchaseReceiptPage({super.key, required this.id});
  final int id;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Comprobante PDF')),
    body: AccountScope(
      child: AsyncContent<Uint8List>(
        load: () =>
            context.read<AndariaApi>().privatePdf('/me/purchases/$id/receipt'),
        builder: (bytes, reload) => Column(
          children: [
            const Padding(
              padding: EdgeInsets.all(12),
              child: Text(
                'Puedes compartir el PDF o guardarlo con la opción Imprimir → Guardar como PDF. El documento muestra el estado al momento de generarlo.',
              ),
            ),
            Expanded(
              child: PdfPreview(
                build: (_) async => bytes,
                pdfFileName: 'andaria-compra-$id.pdf',
                canChangePageFormat: false,
                canChangeOrientation: false,
                canDebug: false,
                allowPrinting: true,
                allowSharing: true,
              ),
            ),
          ],
        ),
      ),
    ),
  );
}
