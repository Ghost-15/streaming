import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../presentation/mobile/pages/home_screen.dart';

/// GoRouter configuration
final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const HomeScreen(),
      ),
    ],
  );
});

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
