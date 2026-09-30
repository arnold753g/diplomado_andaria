import '../core/password_policy.dart';
import '../core/form_validation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../core/theme.dart';
import '../core/google_access.dart';
import '../shared/widgets.dart';

Future<bool> requireTourist(BuildContext context) async {
  final api = context.read<AndariaApi>();
  if (api.signedIn) return true;
  await api.restore();
  if (!context.mounted || ModalRoute.of(context)?.isCurrent == false) {
    return false;
  }
  if (api.signedIn) return true;
  return await openPage<bool>(context, const AuthPage()) ?? false;
}

class AuthPage extends StatefulWidget {
  const AuthPage({super.key});
  @override
  State<AuthPage> createState() => _AuthPageState();
}

class _AuthPageState extends State<AuthPage> {
  final form = GlobalKey<FormState>();
  final email = TextEditingController(),
      password = TextEditingController(),
      first = TextEditingController(),
      last = TextEditingController();
  bool registration = false, busy = false, hidden = true;
  String? error;
  bool registrationEnabled = false;
  String googleClient = '';
  @override
  void initState() {
    super.initState();
    _options();
  }

  Future<void> _options() async {
    try {
      final data = await context.read<AndariaApi>().request(
        '/auth/mobile/options',
      );
      if (mounted) {
        setState(() {
          registrationEnabled = data['registration_enabled'] == true;
          googleClient = !kIsWeb && data['google_enabled'] == true
              ? data['google_server_client_id']?.toString() ?? ''
              : '';
        });
      }
    } catch (_) {}
  }

  Future<void> googleLogin() async {
    setState(() {
      busy = true;
      error = null;
    });
    final api = context.read<AndariaApi>();
    try {
      final token = await GoogleAccess.token(googleClient);
      await api.loginGoogle(token);
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e is ApiFailure
              ? e.message
              : 'No se pudo acceder con Google.';
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          busy = false;
        });
      }
    }
  }

  @override
  void dispose() {
    email.dispose();
    password.dispose();
    first.dispose();
    last.dispose();
    super.dispose();
  }

  Future<void> submit() async {
    if (!form.currentState!.validate()) return;
    setState(() {
      busy = true;
      error = null;
    });
    final api = context.read<AndariaApi>();
    try {
      if (registration) {
        await api.register({
          'email': email.text.trim(),
          'password': password.text,
          'first_name': first.text.trim(),
          'last_name': last.text.trim(),
        });
        if (mounted) {
          setState(() {
            registration = false;
          });
        }
      }
      await api.login(email.text, password.text);
      if (mounted) Navigator.pop(context, true);
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e is ApiFailure ? e.message : 'No se pudo iniciar sesión.';
        });
      }
    } finally {
      if (mounted) {
        setState(() {
          busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(),
    body: AutofillGroup(
      child: Form(
        key: form,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(24, 8, 24, 32),
          children: [
            Align(
              alignment: Alignment.centerLeft,
              child: Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: lime,
                  borderRadius: BorderRadius.circular(18),
                ),
                child: const Icon(Icons.explore_outlined, size: 34),
              ),
            ),
            const SizedBox(height: 28),
            Text(
              registration
                  ? 'Tu próxima aventura\nempieza aquí.'
                  : 'Qué bueno\nverte de nuevo.',
              style: Theme.of(context).textTheme.headlineLarge,
            ),
            const SizedBox(height: 12),
            Text(
              registration
                  ? 'Crea tu cuenta de turista en Andaria.'
                  : 'Ingresa para guardar lugares y gestionar tus compras.',
              style: const TextStyle(color: muted),
            ),
            const SizedBox(height: 28),
            if (registration) ...[
              TextFormField(
                enabled: !busy,
                controller: first,
                maxLength: 100,
                decoration: const InputDecoration(
                  labelText: 'Nombres',
                  counterText: '',
                ),
                textCapitalization: TextCapitalization.words,
                autofillHints: const [AutofillHints.givenName],
                validator: nameError,
              ),
              const SizedBox(height: 14),
              TextFormField(
                enabled: !busy,
                controller: last,
                maxLength: 100,
                decoration: const InputDecoration(
                  labelText: 'Apellidos',
                  counterText: '',
                ),
                textCapitalization: TextCapitalization.words,
                autofillHints: const [AutofillHints.familyName],
                validator: nameError,
              ),
              const SizedBox(height: 14),
            ],
            TextFormField(
              enabled: !busy,
              controller: email,
              maxLength: 320,
              decoration: const InputDecoration(
                labelText: 'Correo electrónico',
                counterText: '',
              ),
              keyboardType: TextInputType.emailAddress,
              autofillHints: const [AutofillHints.email],
              autocorrect: false,
              validator: (v) =>
                  RegExp(
                    r'^[^\s@]+@[^\s@]+\.[^\s@]+$',
                  ).hasMatch(v?.trim() ?? '')
                  ? null
                  : 'Ingresa un correo válido.',
            ),
            const SizedBox(height: 14),
            TextFormField(
              enabled: !busy,
              controller: password,
              obscureText: hidden,
              autofillHints: [
                registration
                    ? AutofillHints.newPassword
                    : AutofillHints.password,
              ],
              autocorrect: false,
              enableSuggestions: false,
              decoration: InputDecoration(
                labelText: 'Contraseña',
                suffixIcon: IconButton(
                  tooltip: hidden ? 'Mostrar contraseña' : 'Ocultar contraseña',
                  icon: Icon(
                    hidden
                        ? Icons.visibility_outlined
                        : Icons.visibility_off_outlined,
                  ),
                  onPressed: () => setState(() {
                    hidden = !hidden;
                  }),
                ),
              ),
              validator: (v) {
                if ((v ?? '').isEmpty) return 'Ingresa tu contraseña.';
                if (registration) return passwordError(v);
                return null;
              },
            ),
            if (registration)
              const Padding(
                padding: EdgeInsets.only(top: 10),
                child: Text(
                  'Te recomendamos una contraseña de al menos 12 letras o números, sin espacios al inicio ni al final.',
                  style: TextStyle(color: muted),
                ),
              ),
            if (error != null) Notice(error!, warning: true),
            const SizedBox(height: 24),
            BusyButton(
              label: registration
                  ? 'Crear cuenta y continuar'
                  : 'Iniciar sesión',
              busy: busy,
              onPressed: submit,
            ),
            if (registrationEnabled)
              TextButton(
                onPressed: busy
                    ? null
                    : () => setState(() {
                        registration = !registration;
                        error = null;
                      }),
                child: Text(
                  registration ? 'Ya tengo una cuenta' : 'Crear una cuenta',
                ),
              ),
            if (googleClient.isNotEmpty) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton(
                  onPressed: busy ? null : googleLogin,
                  child: const Text('Continuar con Google'),
                ),
              ),
            ],
            const SizedBox(height: 20),
            const Text(
              'También puedes explorar el catálogo sin una cuenta.',
              textAlign: TextAlign.center,
              style: TextStyle(color: muted),
            ),
          ],
        ),
      ),
    ),
  );
}
