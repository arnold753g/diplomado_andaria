import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import 'widgets.dart';

// A private form/view belongs to the account that opened it. A later login
// cannot reuse fields or images from the previous account's screen.
class AccountScope extends StatefulWidget {
  const AccountScope({super.key, required this.child});
  final Widget child;
  @override
  State<AccountScope> createState() => _AccountScopeState();
}

class _AccountScopeState extends State<AccountScope> {
  Object? accountId;
  @override
  void initState() {
    super.initState();
    accountId = context.read<AndariaApi>().user?['id'];
  }

  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    if (api.signedIn && accountId != null && accountId == api.user?['id']) {
      return widget.child;
    }
    return EmptyView(
      icon: Icons.lock_outline,
      title: 'La sesión cambió',
      description: 'Regresa e inicia sesión para continuar con tu cuenta.',
      action: 'Volver',
      onAction: () => Navigator.of(context).maybePop(),
    );
  }
}
