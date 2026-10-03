import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:provider/provider.dart';
import 'core/api.dart';
import 'core/theme.dart';
import 'features/shell.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initializeDateFormatting('es_BO');
  LicenseRegistry.addLicense(() async* {
    final text = await rootBundle.loadString('assets/fonts/OFL.txt');
    yield LicenseEntryWithLineBreaks(['Outfit'], text);
  });
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      systemNavigationBarColor: Colors.white,
      systemNavigationBarIconBrightness: Brightness.dark,
    ),
  );
  final api = AndariaApi();
  runApp(AndariaApp(api: api));
  unawaited(api.restore());
}

class AndariaApp extends StatelessWidget {
  const AndariaApp({super.key, required this.api});
  final AndariaApi api;
  @override
  Widget build(BuildContext context) => ChangeNotifierProvider.value(
    value: api,
    child: MaterialApp(
      title: 'Andaria',
      debugShowCheckedModeBanner: false,
      theme: andariaTheme,
      locale: const Locale('es', 'BO'),
      supportedLocales: const [Locale('es', 'BO')],
      localizationsDelegates: GlobalMaterialLocalizations.delegates,
      home: const AppShell(),
    ),
  );
}
