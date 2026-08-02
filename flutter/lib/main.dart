import 'package:flutter/material.dart';
import 'package:flutter_web_plugins/url_strategy.dart';
import 'package:provider/provider.dart';

import 'api/repositories/admin_repository.dart';
import 'api/repositories/favorite_repository.dart';
import 'api/repositories/playlist_repository.dart';
import 'api/repositories/stream_repository.dart';
import 'config/router.dart';
import 'config/theme.dart';
import 'services/api_service.dart';
import 'notifiers/admin_notifier.dart';
import 'notifiers/audio_notifier.dart';
import 'notifiers/broadcaster_notifier.dart';
import 'notifiers/favorite_notifier.dart';
import 'notifiers/playlist_notifier.dart';
import 'notifiers/session_notifier.dart';
import 'notifiers/stream_notifier.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  usePathUrlStrategy();

  final sessionNotifier = SessionNotifier();
  try { await sessionNotifier.init(); } catch (_) {}
  ApiService.onUnauthorized = sessionNotifier.logout;
  final router = buildRouter(sessionNotifier);

  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: sessionNotifier),
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
      ],
      child: StreamPulseApp(router: router),
    ),
  );
}

class StreamPulseApp extends StatelessWidget {
  final dynamic router;
  const StreamPulseApp({super.key, required this.router});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'StreamPulse',
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
