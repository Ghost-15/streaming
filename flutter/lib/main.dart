import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'api/repositories/stream_repository.dart';
import 'config/router.dart';
import 'config/theme.dart';
import 'notifiers/audio_notifier.dart';
import 'notifiers/broadcaster_notifier.dart';
import 'notifiers/session_notifier.dart';
import 'notifiers/stream_notifier.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => SessionNotifier()),
        ChangeNotifierProvider(create: (_) => AudioNotifier()),
        ChangeNotifierProvider(
          create: (_) => StreamNotifier(const StreamRepository()),
        ),
        ChangeNotifierProvider(
          create: (_) => BroadcasterNotifier(const StreamRepository()),
        ),
      ],
      child: const StreamPulseApp(),
    ),
  );
}

class StreamPulseApp extends StatelessWidget {
  const StreamPulseApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'StreamPulse',
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
