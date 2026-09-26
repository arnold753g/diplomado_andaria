import 'package:google_sign_in/google_sign_in.dart';
import 'api.dart';

class GoogleAccess {
  static String? _initializedClient;
  static Future<String> token(String serverClientId) async {
    try {
      final google = GoogleSignIn.instance;
      if (_initializedClient != serverClientId) {
        await google.initialize(serverClientId: serverClientId);
        _initializedClient = serverClientId;
      }
      final account = await google.authenticate();
      final token = account.authentication.idToken;
      if (token == null) {
        throw const ApiFailure('Google no entregó una identidad válida.');
      }
      return token;
    } on GoogleSignInException catch (e) {
      if (e.code == GoogleSignInExceptionCode.canceled) {
        throw const ApiFailure('Se canceló el acceso con Google.');
      }
      if (e.code == GoogleSignInExceptionCode.clientConfigurationError) {
        throw const ApiFailure(
          'El acceso con Google necesita configurar esta aplicación Android.',
        );
      }
      throw const ApiFailure(
        'No se pudo acceder con Google. Inténtalo nuevamente.',
      );
    }
  }
}
