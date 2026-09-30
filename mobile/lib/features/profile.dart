import '../shared/account_scope.dart';
import '../core/password_policy.dart';
import '../core/form_validation.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import '../core/theme.dart';
import '../shared/widgets.dart';
import 'auth.dart';

class ProfilePage extends StatelessWidget {
  const ProfilePage({super.key});
  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    return Scaffold(
      appBar: AppBar(
        title: const Text('Tu perfil'),
        actions: [
          IconButton(
            tooltip: 'Acerca de Andaria',
            icon: const Icon(Icons.info_outline),
            onPressed: () => showAboutDialog(
              context: context,
              applicationName: 'Andaria',
              applicationVersion: '0.1.0',
              applicationIcon: const Icon(Icons.explore_outlined, size: 40),
              children: const [
                Text('Experiencias y lugares de Bolivia para turistas.'),
              ],
            ),
          ),
        ],
      ),
      body: !api.signedIn
          ? EmptyView(
              icon: Icons.person_outline,
              title: 'Viaja con Andaria',
              description:
                  api.sessionNotice ??
                  'Inicia sesión para guardar lugares y gestionar tus compras.',
              action: 'Iniciar sesión',
              onAction: () => requireTourist(context),
            )
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                const Align(
                  alignment: Alignment.centerLeft,
                  child: CircleAvatar(
                    radius: 34,
                    backgroundColor: lime,
                    child: Icon(Icons.person_outline, color: ink, size: 36),
                  ),
                ),
                const SizedBox(height: 20),
                Text(
                  api.fullName,
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                Text(
                  api.user!['email'].toString(),
                  style: const TextStyle(color: muted),
                ),
                const SizedBox(height: 24),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: const Icon(Icons.badge_outlined),
                  title: const Text('Mis datos'),
                  subtitle: const Text('Nombre, teléfono y documento'),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => openPage(context, const EditProfilePage()),
                ),
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: const Icon(Icons.lock_outline),
                  title: Text(
                    api.user!['has_password'] == true
                        ? 'Cambiar contraseña'
                        : 'Crear contraseña',
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => openPage(context, const PasswordPage()),
                ),
                const Divider(),
                const Fact(Icons.language_outlined, 'Español · Bolivia'),
                const Fact(
                  Icons.payments_outlined,
                  'Precios en bolivianos (Bs)',
                ),
                const Fact(
                  Icons.schedule_outlined,
                  'Horarios de Bolivia (UTC−4)',
                ),
                const SizedBox(height: 24),
                OutlinedButton.icon(
                  onPressed: () async {
                    final confirmed = await showDialog<bool>(
                      context: context,
                      builder: (context) => AlertDialog(
                        title: const Text('Cerrar sesión'),
                        content: const Text(
                          'Podrás seguir explorando el catálogo.',
                        ),
                        actions: [
                          TextButton(
                            onPressed: () => Navigator.pop(context, false),
                            child: const Text('Volver'),
                          ),
                          FilledButton(
                            onPressed: () => Navigator.pop(context, true),
                            child: const Text('Cerrar sesión'),
                          ),
                        ],
                      ),
                    );
                    if (confirmed != true) return;
                    try {
                      await api.logout();
                    } catch (e) {
                      if (context.mounted) {
                        message(
                          context,
                          const ApiFailure(
                            'Se cerró la sesión en este dispositivo. No pudimos confirmar la revocación en el servidor.',
                          ),
                        );
                      }
                    }
                  },
                  icon: const Icon(Icons.logout),
                  label: const Text('Cerrar sesión'),
                ),
              ],
            ),
    );
  }
}

class EditProfilePage extends StatefulWidget {
  const EditProfilePage({super.key});
  @override
  State<EditProfilePage> createState() => _EditProfilePageState();
}

class _EditProfilePageState extends State<EditProfilePage> {
  final form = GlobalKey<FormState>();
  final fields = <String, TextEditingController>{};
  bool busy = false;
  String? error;
  static const labels = {
    'first_name': 'Nombres',
    'last_name': 'Apellidos',
    'phone': 'Teléfono',
    'document_number': 'Documento de identidad',
    'nationality': 'Nacionalidad',
  };
  @override
  void initState() {
    super.initState();
    final user = context.read<AndariaApi>().user ?? <String, dynamic>{};
    for (final key in labels.keys) {
      fields[key] = TextEditingController(text: user[key]?.toString() ?? '');
    }
  }

  @override
  void dispose() {
    for (final controller in fields.values) {
      controller.dispose();
    }
    super.dispose();
  }

  Future<void> save() async {
    if (!form.currentState!.validate()) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      await context.read<AndariaApi>().updateProfile(
        fields.map((key, value) => MapEntry(key, value.text.trim())),
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Tus datos se guardaron.')),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e.toString();
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
    appBar: AppBar(title: const Text('Mis datos')),
    body: AccountScope(
      child: Form(
        key: form,
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            const Notice(
              'Estos datos se usan para tus próximas compras. Las compras anteriores conservan los datos registrados al realizarlas.',
            ),
            ...labels.entries.map(
              (entry) => Padding(
                padding: const EdgeInsets.only(bottom: 16),
                child: TextFormField(
                  enabled: !busy,
                  controller: fields[entry.key],
                  decoration: InputDecoration(labelText: entry.value),
                  maxLength: entry.key == 'phone'
                      ? 30
                      : entry.key == 'document_number'
                      ? 40
                      : entry.key == 'nationality'
                      ? 80
                      : 100,
                  keyboardType: entry.key == 'phone'
                      ? TextInputType.phone
                      : TextInputType.text,
                  validator: (v) =>
                      ['first_name', 'last_name'].contains(entry.key)
                      ? nameError(v)
                      : null,
                ),
              ),
            ),
            if (error != null) Notice(error!, warning: true),
            BusyButton(label: 'Guardar cambios', busy: busy, onPressed: save),
          ],
        ),
      ),
    ),
  );
}

class PasswordPage extends StatefulWidget {
  const PasswordPage({super.key});
  @override
  State<PasswordPage> createState() => _PasswordPageState();
}

class _PasswordPageState extends State<PasswordPage> {
  final form = GlobalKey<FormState>();
  final current = TextEditingController(),
      next = TextEditingController(),
      repeat = TextEditingController();
  bool busy = false;
  String? error;
  @override
  void dispose() {
    current.dispose();
    next.dispose();
    repeat.dispose();
    super.dispose();
  }

  Future<void> save() async {
    if (!form.currentState!.validate()) return;
    setState(() {
      busy = true;
      error = null;
    });
    final api = context.read<AndariaApi>();
    try {
      await api.request(
        '/me/password',
        method: 'POST',
        body: {'current_password': current.text, 'new_password': next.text},
        authenticated: true,
      );
      try {
        await api.clear();
      } catch (_) {
        // The confirmed password update revoked every previous server session.
        // clear() hides local private data before deleting persisted credentials.
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Contraseña guardada. Inicia sesión nuevamente.'),
          ),
        );
        Navigator.pop(context);
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          error = e.toString();
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
  Widget build(BuildContext context) {
    final hasPassword =
        context.read<AndariaApi>().user?['has_password'] == true;
    return Scaffold(
      appBar: AppBar(
        title: Text(hasPassword ? 'Cambiar contraseña' : 'Crear contraseña'),
      ),
      body: AccountScope(
        child: Form(
          key: form,
          child: ListView(
            padding: const EdgeInsets.all(20),
            children: [
              if (hasPassword) ...[
                TextFormField(
                  enabled: !busy,
                  controller: current,
                  autofillHints: const [AutofillHints.password],
                  autocorrect: false,
                  enableSuggestions: false,
                  obscureText: true,
                  decoration: const InputDecoration(
                    labelText: 'Contraseña actual',
                  ),
                  validator: (v) =>
                      (v ?? '').isEmpty ? 'Ingresa tu contraseña.' : null,
                ),
                const SizedBox(height: 16),
              ],
              TextFormField(
                enabled: !busy,
                controller: next,
                autofillHints: const [AutofillHints.newPassword],
                autocorrect: false,
                enableSuggestions: false,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Nueva contraseña',
                ),
                validator: passwordError,
              ),
              const SizedBox(height: 16),
              TextFormField(
                enabled: !busy,
                controller: repeat,
                autofillHints: const [AutofillHints.newPassword],
                autocorrect: false,
                enableSuggestions: false,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Repite la nueva contraseña',
                ),
                validator: (v) =>
                    v != next.text ? 'Las contraseñas no coinciden.' : null,
              ),
              const Notice(
                'Al guardar la contraseña se cerrarán todas tus sesiones.',
              ),
              if (error != null) Notice(error!, warning: true),
              const SizedBox(height: 16),
              BusyButton(
                label: 'Guardar contraseña',
                busy: busy,
                onPressed: save,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
