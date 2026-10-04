import 'dart:convert';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';

const tinyPng =
    'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGNYtfsMAAREAjLFENlZAAAAAElFTkSuQmCC';

class Gallery extends ImagePickerPlatform {
  @override
  Future<XFile?> getImageFromSource({
    required ImageSource source,
    ImagePickerOptions options = const ImagePickerOptions(),
  }) async => XFile.fromData(
    base64Decode(tinyPng),
    name: 'pago.png',
    mimeType: 'image/png',
  );
}
