import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';
import 'package:streampulse/api/repositories/admin_repository.dart';
import 'package:streampulse/api/repositories/favorite_repository.dart';
import 'package:streampulse/api/repositories/playlist_repository.dart';
import 'package:streampulse/api/repositories/stream_repository.dart';
import 'package:streampulse/config/router.dart';
import 'package:streampulse/main.dart';
import 'package:streampulse/notifiers/admin_notifier.dart';
import 'package:streampulse/notifiers/audio_notifier.dart';
import 'package:streampulse/notifiers/broadcaster_notifier.dart';
import 'package:streampulse/notifiers/favorite_notifier.dart';
import 'package:streampulse/notifiers/playlist_notifier.dart';
import 'package:streampulse/notifiers/recommendation_notifier.dart';
import 'package:streampulse/notifiers/session_notifier.dart';
import 'package:streampulse/notifiers/stream_notifier.dart';

void main() {
  testWidgets('StreamPulse app smoke test', (WidgetTester tester) async {
    final session = SessionNotifier();
    await tester.pumpWidget(
      MultiProvider(
        providers: [
          ChangeNotifierProvider.value(value: session),
          ChangeNotifierProvider(create: (_) => AudioNotifier()),
          ChangeNotifierProvider(
            create: (_) => StreamNotifier(const StreamRepository()),
          ),
          ChangeNotifierProvider(
            create: (_) => BroadcasterNotifier(const StreamRepository()),
          ),
          ChangeNotifierProvider(
            create: (_) => AdminNotifier(const AdminRepository()),
          ),
          ChangeNotifierProvider(
            create: (_) => PlaylistNotifier(const PlaylistRepository()),
          ),
          ChangeNotifierProvider(
            create: (_) => FavoriteNotifier(const FavoriteRepository()),
          ),
          ChangeNotifierProvider(
            create: (_) => RecommendationNotifier(const StreamRepository()),
          ),
        ],
        child: StreamPulseApp(router: buildRouter(session)),
      ),
    );
    expect(find.text('StreamPulse'), findsWidgets);
  });
}
