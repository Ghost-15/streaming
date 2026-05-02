import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/app.dart';

void main() {
  testWidgets('StreamPulse app smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(child: StreamPulseApp()),
    );
    // L'app démarre sur LoginPage — vérifier que le titre est présent
    expect(find.text('StreamPulse'), findsOneWidget);
  });
}
