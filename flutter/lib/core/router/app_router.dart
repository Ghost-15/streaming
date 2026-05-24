import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../presentation/mobile/pages/home_screen.dart';
import '../../presentation/mobile/pages/audio_player_screen.dart';

/// GoRouter configuration
final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const HomeScreen(),
      ),
      GoRoute(
        path: '/player',
        builder: (context, state) => const AudioPlayerScreen(
          streamId: null,
        ),
      ),
      GoRoute(
        path: '/player/:streamId',
        builder: (context, state) => AudioPlayerScreen(
          streamId: state.pathParameters['streamId'],
        ),
      ),
    ],
  );
});
