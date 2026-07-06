import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../api/models/role.dart';
import '../notifiers/session_notifier.dart';
import '../screens/admin_screen.dart';
import '../screens/audio_player_screen.dart';
import '../screens/broadcaster_screen.dart';
import '../screens/home_screen.dart';
import '../screens/login_screen.dart';

Page<dynamic> buildPage(BuildContext context, GoRouterState state, Widget child) {
  if (kIsWeb) return NoTransitionPage(child: child);
  return MaterialPage(child: child);
}

String _homeForRole(Role role) => switch (role) {
      Role.admin => '/admin',
      Role.diffuseur => '/broadcaster',
      _ => '/',
    };

GoRouter buildRouter(SessionNotifier session) => GoRouter(
      debugLogDiagnostics: true,
      initialLocation: '/',
      refreshListenable: session,
      redirect: (context, state) {
        final authenticated = session.isAuthenticated;
        final loc = state.matchedLocation;

        if (authenticated && loc == '/login') {
          return _homeForRole(session.user!.role);
        }

        const protected = ['/broadcaster', '/admin', '/profile'];
        if (!authenticated && protected.any((p) => loc.startsWith(p))) {
          return '/login';
        }

        return null;
      },
      routes: [
        GoRoute(
          path: '/',
          pageBuilder: (context, state) =>
              buildPage(context, state, const HomeScreen()),
        ),
        GoRoute(
          path: '/login',
          pageBuilder: (context, state) =>
              buildPage(context, state, const LoginScreen()),
        ),
        GoRoute(
          path: '/broadcaster',
          pageBuilder: (context, state) =>
              buildPage(context, state, const BroadcasterScreen()),
        ),
        GoRoute(
          path: '/admin',
          pageBuilder: (context, state) =>
              buildPage(context, state, const AdminScreen()),
        ),
        GoRoute(
          path: '/player',
          pageBuilder: (context, state) =>
              buildPage(context, state, const AudioPlayerScreen()),
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
