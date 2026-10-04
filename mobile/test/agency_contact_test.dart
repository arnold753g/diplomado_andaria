import 'package:andaria_mobile/shared/agency_contact.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:url_launcher_platform_interface/url_launcher_platform_interface.dart';
import 'package:url_launcher_platform_interface/link.dart';

class FakeLauncher extends UrlLauncherPlatform {
  @override
  LinkDelegate? get linkDelegate => null;
  final List<Uri> opened = [];
  bool available = true;
  @override
  Future<bool> launchUrl(String url, LaunchOptions options) async {
    expect(options.mode, PreferredLaunchMode.externalApplication);
    opened.add(Uri.parse(url));
    return available;
  }
}

void main() {
  testWidgets(
    'commercial contact opens dialer and safely encoded email draft',
    (tester) async {
      final previous = UrlLauncherPlatform.instance;
      final fake = FakeLauncher();
      UrlLauncherPlatform.instance = fake;
      addTearDown(() => UrlLauncherPlatform.instance = previous);
      const reference = 'AND-123 & revisión #4';
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: AgencyContact(
              name: 'Agencia',
              phone: '+591 (7111) 2223',
              email: 'ventas@example.test',
              reference: reference,
            ),
          ),
        ),
      );
      await tester.tap(find.text('Abrir teléfono'));
      await tester.pumpAndSettle();
      expect(fake.opened.single.toString(), 'tel:+59171112223');
      await tester.tap(find.text('Redactar correo'));
      await tester.pumpAndSettle();
      expect(fake.opened.last.scheme, 'mailto');
      expect(fake.opened.last.path, 'ventas@example.test');
      expect(fake.opened.last.queryParameters, {
        'subject': 'Consulta sobre la compra $reference',
      });
      fake.available = false;
      await tester.tap(find.text('Redactar correo'));
      await tester.pumpAndSettle();
      expect(
        find.textContaining('Puedes copiar el dato de la agencia.'),
        findsOneWidget,
      );
    },
  );
  testWidgets(
    'ambiguous contact is copyable without offering a wrong dial number',
    (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: AgencyContact(
              name: 'Agencia',
              phone: '71112223 ext. 123',
              email: 'sin correo',
            ),
          ),
        ),
      );
      expect(find.text('71112223 ext. 123'), findsOneWidget);
      expect(find.text('Abrir teléfono'), findsNothing);
      expect(find.text('Redactar correo'), findsNothing);
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: AgencyContact(name: 'Agencia', phone: '', email: ''),
          ),
        ),
      );
      expect(find.text('Contacta a la agencia'), findsNothing);
    },
  );
}
