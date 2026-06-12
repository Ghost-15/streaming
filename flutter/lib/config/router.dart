import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../screens/audio_player_screen.dart';
import '../screens/home_screen.dart';
import '../screens/login_screen.dart';

Page<dynamic> buildPage(BuildContext context, GoRouterState state, Widget child) {
  if (kIsWeb) return NoTransitionPage(child: child);
  return MaterialPage(child: child);
}

final GoRouter router = GoRouter(
  debugLogDiagnostics: true,
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      pageBuilder: (context, state) => buildPage(context, state, const HomeScreen()),
    ),
    GoRoute(
      path: '/login',
      pageBuilder: (context, state) => buildPage(context, state, const LoginScreen()),
    ),
    GoRoute(
      path: '/player',
      pageBuilder: (context, state) => buildPage(context, state, const AudioPlayerScreen()),
      routes: [
        GoRoute(
          path: ':streamId',
          pageBuilder: (context, state) => buildPage(
            context,
            state,
            AudioPlayerScreen(streamId: state.pathParameters['streamId']),
          ),
        ),
      ],
    ),
  ],
);
