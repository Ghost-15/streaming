import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:streampulse/screens/login_screen.dart';
import 'package:streampulse/screens/register_screen.dart';
import 'package:streampulse/widgets/page_header.dart';

// These tests lock in the screen reader contract. They are deliberately generic:
// a new icon button or a new page title fails them until it is labelled.

Widget _wrap(Widget child) => MaterialApp(home: child);

void main() {
  testWidgets('PageHeader exposes its title as a heading', (tester) async {
    final handle = tester.ensureSemantics();

    await tester.pumpWidget(
      _wrap(
        const Scaffold(
          body: PageHeader(icon: Icons.library_music, title: 'Bibliothèque'),
        ),
      ),
    );

    final semantics = tester.getSemantics(find.text('Bibliothèque'));
    expect(
      semantics.flagsCollection.isHeader,
      isTrue,
      reason: 'a screen title must be announced as a heading so a screen '
          'reader user can jump between sections',
    );

    handle.dispose();
  });

  testWidgets('PageHeader hides its decorative icon from screen readers', (
    tester,
  ) async {
    final handle = tester.ensureSemantics();

    await tester.pumpWidget(
      _wrap(
        const Scaffold(
          body: PageHeader(icon: Icons.library_music, title: 'Bibliothèque'),
        ),
      ),
    );

    // The icon repeats the title, so it must not add a second announcement.
    expect(find.bySemanticsLabel('Bibliothèque'), findsOneWidget);

    handle.dispose();
  });

  testWidgets('every icon button of the login screen is labelled', (
    tester,
  ) async {
    await tester.pumpWidget(_wrap(const LoginScreen()));
    await tester.pump();

    _expectEveryIconButtonLabelled(tester, 'LoginScreen');
  });

  testWidgets('every icon button of the register screen is labelled', (
    tester,
  ) async {
    await tester.pumpWidget(_wrap(const RegisterScreen()));
    await tester.pump();

    _expectEveryIconButtonLabelled(tester, 'RegisterScreen');
  });

  testWidgets('the password toggle announces what it will do', (tester) async {
    await tester.pumpWidget(_wrap(const LoginScreen()));
    await tester.pump();

    expect(find.byTooltip('Afficher le mot de passe'), findsOneWidget);

    await tester.tap(find.byTooltip('Afficher le mot de passe'));
    await tester.pump();

    expect(
      find.byTooltip('Masquer le mot de passe'),
      findsOneWidget,
      reason: 'the label must follow the state, not describe the icon',
    );
  });

  testWidgets('form fields carry a visible label', (tester) async {
    await tester.pumpWidget(_wrap(const LoginScreen()));
    await tester.pump();

    for (final label in ['Email', 'Mot de passe']) {
      expect(
        find.text(label),
        findsOneWidget,
        reason: 'the field $label must keep its label for assistive technology',
      );
    }
  });
}

void _expectEveryIconButtonLabelled(WidgetTester tester, String screen) {
  final buttons = tester.widgetList<IconButton>(find.byType(IconButton));
  expect(
    buttons,
    isNotEmpty,
    reason: '$screen was expected to render at least one icon button',
  );

  for (final button in buttons) {
    final label = button.tooltip;
    expect(
      label != null && label.trim().isNotEmpty,
      isTrue,
      reason: '$screen has an icon button without a tooltip: an icon alone '
          'is announced as "button" and gives no clue about its action',
    );
  }
}
