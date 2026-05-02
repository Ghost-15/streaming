// GoRouter configuration.
// Redirects based on auth state + role:
//   - Not logged in → /login
//   - admin on web  → /admin/dashboard
//   - others        → /streams
// Sprint 1 — US-002.
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../presentation/mobile/login_page.dart';
import '../../presentation/mobile/stream_page.dart';
import '../../presentation/providers/auth_provider.dart';
import '../../presentation/web/admin_dashboard_page.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final user = ref.watch(currentUserProvider);

  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      final loggedIn = user != null;
      final isLogin = state.matchedLocation == '/login';

      if (!loggedIn) return isLogin ? null : '/login';

      if (user.isAdmin && kIsWeb) return '/admin/dashboard';

      return isLogin ? '/streams' : null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/streams', builder: (context, state) => const StreamPage()),
      GoRoute(
        path: '/admin/dashboard',
        builder: (context, state) => const AdminDashboardPage(),
      ),
      // TODO Sprint 2 — US-009: /streams/:id/player
      // TODO Sprint 2 — US-007: /playlists
      // TODO Sprint 3 — US-013: /admin/users
    ],
  );
});
