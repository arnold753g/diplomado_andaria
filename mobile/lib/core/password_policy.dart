import 'dart:convert';

// Match the backend's bcrypt limits without exposing byte counts in forms.
String? passwordError(String? value) {
  final password = value ?? '';
  if (password != password.trim()) {
    return 'Evita espacios al inicio o al final de la contraseña.';
  }
  final length = utf8.encode(password).length;
  if (length < 12) {
    return 'La contraseña es demasiado corta. Usa al menos 12 letras o números.';
  }
  if (length > 72) {
    return 'La contraseña es demasiado larga. Reduce su longitud.';
  }
  return null;
}
