import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:streampulse/api/repositories/stream_repository.dart';
import 'package:streampulse/main.dart';
import 'package:streampulse/notifiers/audio_notifier.dart';
import 'package:streampulse/notifiers/session_notifier.dart';
import 'package:streampulse/notifiers/stream_notifier.dart';

void main() {
  testWidgets('StreamPulse app smoke test', (WidgetTester tester) async {
    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider(create: (_) => SessionNotifier()),
          ChangeNotifierProvider(create: (_) => AudioNotifier()),
          ChangeNotifierProvider(
            create: (_) => StreamNotifier(const StreamRepository()),
          ),
        ],
        child: const StreamPulseApp(),
      ),
    );
    expect(find.text('StreamPulse'), findsWidgets);
  });
}
