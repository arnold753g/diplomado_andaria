import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../core/api.dart';
import 'catalog.dart';
import 'purchases.dart';
import 'profile.dart';

class AppShell extends StatefulWidget {
  const AppShell({super.key});
  @override
  State<AppShell> createState() => _AppShellState();
}

class _AppShellState extends State<AppShell> with WidgetsBindingObserver {
  int index = 0;
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<AndariaApi>().restore();
    }
  }

  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    return Scaffold(
      body: IndexedStack(
        index: index,
        children: [
          const ExplorePage(),
          index == 1
              ? FavoritesPage(key: ValueKey(api.user?['id']))
              : const SizedBox.shrink(),
          index == 2
              ? PurchasesPage(key: ValueKey(api.user?['id']))
              : const SizedBox.shrink(),
          const ProfilePage(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: index,
        onDestinationSelected: (value) => setState(() {
          index = value;
        }),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.explore_outlined),
            selectedIcon: Icon(Icons.explore),
            label: 'Explorar',
          ),
          NavigationDestination(
            icon: Icon(Icons.favorite_border),
            selectedIcon: Icon(Icons.favorite),
            label: 'Favoritos',
          ),
          NavigationDestination(
            icon: Icon(Icons.confirmation_number_outlined),
            selectedIcon: Icon(Icons.confirmation_number),
            label: 'Mis compras',
          ),
          NavigationDestination(
            icon: Icon(Icons.person_outline),
            selectedIcon: Icon(Icons.person),
            label: 'Perfil',
          ),
        ],
      ),
    );
  }
}
