// Go's validator counts Unicode code points for the backend's name bounds.
String? nameError(String? value) {
  final length = (value ?? '').trim().runes.length;
  if (length < 2) return 'Ingresa al menos 2 caracteres.';
  if (length > 100) return 'Usa como máximo 100 caracteres.';
  return null;
}
