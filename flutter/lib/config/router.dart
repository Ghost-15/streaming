import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../api/models/role.dart';
import '../notifiers/session_notifier.dart';
import '../screens/admin_screen.dart';
import '../screens/audio_player_screen.dart';
import '../screens/broadcaster_screen.dart';
import '../screens/home_screen.dart';
import '../screens/library_screen.dart';
import '../screens/login_screen.dart';
import '../screens/register_screen.dart';
import '../widgets/nav_shell.dart';

Page<dynamic> _page(BuildContext context, GoRouterState state, Widget child) {
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

    if (authenticated && (loc == '/login' || loc == '/register')) {
      return _homeForRole(session.user!.role);
    }

    const protected = ['/broadcaster', '/admin', '/library'];
    if (!authenticated && protected.any((p) => loc.startsWith(p))) {
      return '/login';
    }

    return null;
  },
  routes: [
    StatefulShellRoute.indexedStack(
      builder: (context, state, shell) =>
          ScaffoldWithNavBar(navigationShell: shell),
      branches: [
        // Branch 0 — Accueil
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/',
              pageBuilder: (ctx, state) =>
                  _page(ctx, state, const HomeScreen()),
            ),
          ],
        ),
        // Branch 1 — Live (auditeurs)
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/player',
              pageBuilder: (ctx, state) =>
                  _page(ctx, state, const AudioPlayerScreen()),
            ),
          ],
        ),
        // Branch 2 — Studio (diffuseurs/admin)
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/broadcaster',
              pageBuilder: (ctx, state) =>
                  _page(ctx, state, const BroadcasterScreen()),
            ),
          ],
        ),
        // Branch 3 — Bibliothèque
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/library',
              pageBuilder: (ctx, state) =>
                  _page(ctx, state, const LibraryScreen()),
            ),
          ],
        ),
      ],
    ),

    GoRoute(
      path: '/login',
      pageBuilder: (ctx, state) => _page(ctx, state, const LoginScreen()),
    ),
    GoRoute(
      path: '/register',
      pageBuilder: (ctx, state) => _page(ctx, state, const RegisterScreen()),
    ),
    GoRoute(
      path: '/admin',
      pageBuilder: (ctx, state) => _page(ctx, state, const AdminScreen()),
    ),
  ],
);
