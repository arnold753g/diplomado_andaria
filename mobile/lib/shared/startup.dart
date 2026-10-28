import 'package:flutter/material.dart';
import 'brand.dart';

class AndariaStartup extends StatefulWidget {
  const AndariaStartup({
    super.key,
    required this.initialize,
    required this.child,
  });
  final Future<void> Function() initialize;
  final Widget child;

  @override
  State<AndariaStartup> createState() => _AndariaStartupState();
}

class _AndariaStartupState extends State<AndariaStartup> {
  late Future<void> ready;

  @override
  void initState() {
    super.initState();
    ready = Future<void>.sync(widget.initialize);
  }

  @override
  Widget build(BuildContext context) => FutureBuilder<void>(
    future: ready,
    builder: (context, snapshot) {
      if (snapshot.connectionState == ConnectionState.done &&
          !snapshot.hasError) {
        return widget.child;
      }
      return Scaffold(
        body: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              child:
                  snapshot.connectionState == ConnectionState.done &&
                      snapshot.hasError
                  ? Padding(
                      padding: const EdgeInsets.all(24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const AndariaLogo(size: 72),
                          const SizedBox(height: 24),
                          const Text(
                            'No pudimos iniciar la aplicación',
                            textAlign: TextAlign.center,
                          ),
                          const SizedBox(height: 16),
                          FilledButton(
                            onPressed: () => setState(() {
                              ready = Future<void>.sync(widget.initialize);
                            }),
                            child: const Text('Volver a intentar'),
                          ),
                        ],
                      ),
                    )
                  : const AndariaLoading(
                      label: 'Preparando tu próxima experiencia',
                    ),
            ),
          ),
        ),
      );
    },
  );
}
