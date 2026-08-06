import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../api/models/playlist_model.dart';
import '../api/models/role.dart';
import '../api/models/stream_model.dart';
import '../api/models/track_model.dart';
import '../notifiers/audio_notifier.dart';
import '../notifiers/favorite_notifier.dart';
import '../notifiers/playlist_notifier.dart';
import '../notifiers/session_notifier.dart';
import '../notifiers/stream_notifier.dart';
import '../widgets/stream_card.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<StreamNotifier>().loadActive();
      if (context.read<SessionNotifier>().isAuthenticated) {
        context.read<PlaylistNotifier>().load();
        context.read<FavoriteNotifier>().load();
      }
    });
  }

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 12) return 'Bonjour';
    if (h < 18) return 'Bon après-midi';
    return 'Bonsoir';
  }

  void _play(BuildContext context, StreamModel stream) {
    context.read<AudioNotifier>().playStream(stream);
    context.go('/player');
  }

  Future<void> _refresh() async {
    final isAuth = context.read<SessionNotifier>().isAuthenticated;
    await context.read<StreamNotifier>().loadActive();
    if (!mounted) return;
    if (isAuth) {
      context.read<PlaylistNotifier>().load();
      context.read<FavoriteNotifier>().load();
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();
    final streams = context.watch<StreamNotifier>();
    final playlists = context.watch<PlaylistNotifier>();
    final favorites = context.watch<FavoriteNotifier>();

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: CustomScrollView(
          slivers: [
            _HomeHeader(session: session, greeting: _greeting()),

            // ── Playlists (connecté) ─────────────────────────────────────
            if (session.isAuthenticated && playlists.playlists.isNotEmpty) ...[
              const _SectionHeader('Mes playlists'),
              SliverPadding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                sliver: SliverGrid.builder(
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    crossAxisSpacing: 12,
                    mainAxisSpacing: 12,
                    childAspectRatio: 1.25,
                  ),
                  itemCount: playlists.playlists.length.clamp(0, 4),
                  itemBuilder: (_, i) =>
                      _PlaylistCard(playlist: playlists.playlists[i]),
                ),
              ),
            ],

            // ── Streams ──────────────────────────────────────────────────
            if (streams.isLoading)
              const _LoadingSliver()
            else if (streams.streams.isEmpty)
              const SliverFillRemaining(child: _EmptyState())
            else ...[
              if (!session.isAuthenticated) ...[
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(20, 8, 20, 0),
                    child: _FeaturedCard(
                      stream: streams.streams.first,
                      onPlay: () => _play(context, streams.streams.first),
                    ),
                  ),
                ),
                if (streams.streams.length > 1) ...[
                  const _SectionHeader('En direct maintenant'),
                  SliverToBoxAdapter(
                    child: SizedBox(
                      height: 210,
                      child: ListView.separated(
                        scrollDirection: Axis.horizontal,
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        itemCount: streams.streams.length,
                        separatorBuilder: (_, _) => const SizedBox(width: 14),
                        itemBuilder: (_, i) => StreamCard(
                          stream: streams.streams[i],
                          compact: true,
                          onPlay: () => _play(context, streams.streams[i]),
                        ),
                      ),
                    ),
                  ),
                ],
              ],
              const _SectionHeader('En direct'),
              SliverPadding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                sliver: SliverList.separated(
                  separatorBuilder: (_, _) => const SizedBox(height: 10),
                  itemCount: streams.streams.length,
                  itemBuilder: (_, i) => StreamCard(
                    stream: streams.streams[i],
                    onPlay: () => _play(context, streams.streams[i]),
                  ),
                ),
              ),
            ],

            // ── Favoris (connecté, après les lives) ──────────────────────
            if (session.isAuthenticated && favorites.tracks.isNotEmpty) ...[
              const _SectionHeader('Mes favoris'),
              SliverToBoxAdapter(
                child: SizedBox(
                  height: 64,
                  child: ListView.separated(
                    scrollDirection: Axis.horizontal,
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    itemCount: favorites.tracks.length,
                    separatorBuilder: (_, _) => const SizedBox(width: 10),
                    itemBuilder: (_, i) =>
                        _FavoriteChip(track: favorites.tracks[i]),
                  ),
                ),
              ),
            ],

            if (!session.isAuthenticated)
              SliverToBoxAdapter(
                child: _LoginBanner(onTap: () => context.go('/register')),
              ),
            const SliverPadding(padding: EdgeInsets.only(bottom: 140)),
          ],
        ),
      ),
    );
  }
}

// ── Home header — glassmorphism card combining logo + greeting ────────────────

class _HomeHeader extends StatelessWidget {
  final SessionNotifier session;
  final String greeting;
  const _HomeHeader({required this.session, required this.greeting});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final user = session.user;
    final topPadding = MediaQuery.of(context).padding.top;
    final name = user?.firstName;

    return SliverToBoxAdapter(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(height: topPadding),
          // ── Top bar: logo + actions ─────────────────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 14, 12, 0),
            child: Row(
              children: [
                Container(
                  width: 30,
                  height: 30,
                  decoration: BoxDecoration(
                    color: cs.primary,
                    borderRadius: BorderRadius.circular(9),
                  ),
                  child: Icon(
                    Icons.radio_rounded,
                    color: cs.onPrimary,
                    size: 17,
                  ),
                ),
                const SizedBox(width: 9),
                Text(
                  'StreamPulse',
                  style: tt.titleMedium?.copyWith(
                    fontWeight: FontWeight.w800,
                    letterSpacing: -0.5,
                  ),
                ),
                const Spacer(),
                if (session.isAuthenticated && user != null) ...[
                  if (user.role == Role.admin)
                    IconButton(
                      icon: const Icon(
                        Icons.admin_panel_settings_outlined,
                        size: 20,
                      ),
                      tooltip: 'Administration',
                      color: cs.onSurfaceVariant,
                      visualDensity: VisualDensity.compact,
                      onPressed: () => context.go('/admin'),
                    ),
                  if (user.role == Role.diffuseur)
                    IconButton(
                      icon: const Icon(Icons.mic_none_rounded, size: 20),
                      tooltip: 'Studio',
                      color: cs.onSurfaceVariant,
                      visualDensity: VisualDensity.compact,
                      onPressed: () => context.go('/broadcaster'),
                    ),
                  GestureDetector(
                    onTap: () => _showAccountMenu(context, session),
                    child: CircleAvatar(
                      radius: 15,
                      backgroundColor: cs.primaryContainer,
                      child: Text(
                        _initial(user.firstName, user.email),
                        style: tt.labelMedium?.copyWith(
                          color: cs.onPrimaryContainer,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                ] else ...[
                  GestureDetector(
                    onTap: () => context.go('/login'),
                    child: Container(
                      width: 32,
                      height: 32,
                      decoration: BoxDecoration(
                        color: cs.surfaceContainerHighest,
                        shape: BoxShape.circle,
                        border: Border.all(color: cs.outlineVariant, width: 1),
                      ),
                      child: Icon(
                        Icons.person_outline_rounded,
                        size: 17,
                        color: cs.onSurfaceVariant,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                ],
              ],
            ),
          ),
          // ── Greeting ────────────────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 20, 20, 4),
            child: Text(
              name != null && name.isNotEmpty ? '$greeting, $name' : greeting,
              style: tt.headlineMedium?.copyWith(
                fontWeight: FontWeight.w800,
                letterSpacing: -0.5,
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 0, 20, 16),
            child: Text(
              'Découvre les streams en direct',
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
            ),
          ),
          Divider(
            height: 1,
            thickness: 0.5,
            color: cs.outlineVariant.withValues(alpha: 0.5),
          ),
        ],
      ),
    );
  }

  String _initial(String firstName, String email) {
    if (firstName.isNotEmpty) return firstName[0].toUpperCase();
    if (email.isNotEmpty) return email[0].toUpperCase();
    return '?';
  }

  void _showAccountMenu(BuildContext context, SessionNotifier session) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final user = session.user!;

    showModalBottomSheet<void>(
      context: context,
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 10),
              child: Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: cs.outlineVariant,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 4, 20, 16),
              child: Row(
                children: [
                  CircleAvatar(
                    radius: 24,
                    backgroundColor: cs.primaryContainer,
                    child: Text(
                      _initial(user.firstName, user.email),
                      style: tt.titleMedium?.copyWith(
                        color: cs.onPrimaryContainer,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          user.fullName.trim().isNotEmpty
                              ? user.fullName
                              : user.email,
                          style: tt.titleSmall?.copyWith(
                            fontWeight: FontWeight.w700,
                          ),
                        ),
                        Text(
                          user.email,
                          style: tt.bodySmall?.copyWith(
                            color: cs.onSurfaceVariant,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 2,
                          ),
                          decoration: BoxDecoration(
                            color: cs.primaryContainer.withValues(alpha: 0.4),
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            _roleLabel(user.role),
                            style: TextStyle(
                              fontSize: 11,
                              color: cs.primary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Divider(height: 1, color: cs.outlineVariant),
            ListTile(
              leading: Icon(Icons.logout_rounded, color: cs.error, size: 20),
              title: Text(
                'Se déconnecter',
                style: tt.bodyMedium?.copyWith(color: cs.error),
              ),
              onTap: () {
                Navigator.pop(context);
                session.logout();
              },
            ),
            const SizedBox(height: 8),
          ],
        ),
      ),
    );
  }

  String _roleLabel(Role role) => switch (role) {
    Role.admin => 'Administrateur',
    Role.diffuseur => 'Diffuseur',
    _ => 'Auditeur',
  };
}

// ── Section header ────────────────────────────────────────────────────────────

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader(this.title);

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 32, 20, 14),
        child: Row(
          children: [
            Container(
              width: 3,
              height: 18,
              decoration: BoxDecoration(
                color: cs.primary,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(width: 10),
            Text(
              title,
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w700,
                letterSpacing: -0.3,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Featured hero card ────────────────────────────────────────────────────────

class _FeaturedCard extends StatelessWidget {
  final StreamModel stream;
  final VoidCallback onPlay;
  const _FeaturedCard({required this.stream, required this.onPlay});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final accent = _streamAccent(stream.id);

    return GestureDetector(
      onTap: onPlay,
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: accent.withValues(alpha: 0.28), width: 1),
          boxShadow: [
            BoxShadow(
              color: accent.withValues(alpha: 0.22),
              blurRadius: 24,
              spreadRadius: -4,
              offset: const Offset(0, 6),
            ),
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.3),
              blurRadius: 14,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(19),
          child: SizedBox(
            height: 220,
            child: Stack(
              fit: StackFit.expand,
              children: [
                // Gradient background
                Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        accent,
                        accent.withValues(alpha: 0.6),
                        cs.surfaceContainerHigh,
                      ],
                      stops: const [0.0, 0.5, 1.0],
                    ),
                  ),
                ),
                // Decorative waveform icon
                Positioned(
                  right: -20,
                  top: -20,
                  child: Icon(
                    Icons.radio_rounded,
                    size: 180,
                    color: Colors.white.withValues(alpha: 0.06),
                  ),
                ),
                // Content
                Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // LIVE badge
                      if (stream.isLive) const _LiveBadge(large: true),
                      const Spacer(),
                      // Title
                      Text(
                        stream.title,
                        style: tt.headlineSmall?.copyWith(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                          shadows: [
                            const Shadow(color: Colors.black54, blurRadius: 8),
                          ],
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'par ${stream.broadcasterName}',
                        style: tt.bodyMedium?.copyWith(
                          color: Colors.white.withValues(alpha: 0.8),
                        ),
                      ),
                      const SizedBox(height: 16),
                      // Bottom row
                      Row(
                        children: [
                          _ListenerCount(
                            count: stream.listenerCount,
                            light: true,
                          ),
                          const Spacer(),
                          // Play button
                          Container(
                            width: 52,
                            height: 52,
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(26),
                              boxShadow: [
                                const BoxShadow(
                                  color: Colors.black38,
                                  blurRadius: 12,
                                  offset: Offset(0, 4),
                                ),
                              ],
                            ),
                            child: Icon(
                              Icons.play_arrow_rounded,
                              color: accent,
                              size: 30,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ── Loading ───────────────────────────────────────────────────────────────────

class _LoadingSliver extends StatelessWidget {
  const _LoadingSliver();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return SliverToBoxAdapter(
      child: SizedBox(
        height: 300,
        child: Center(child: CircularProgressIndicator(color: cs.primary)),
      ),
    );
  }
}

// ── Empty state ───────────────────────────────────────────────────────────────

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(40),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: cs.surfaceContainerHigh,
                borderRadius: BorderRadius.circular(40),
              ),
              child: Icon(
                Icons.radio_outlined,
                size: 40,
                color: cs.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 20),
            Text(
              'Aucun stream en direct',
              style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w700),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'Reviens plus tard ou lance ton propre stream.',
              style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

// ── Login banner ──────────────────────────────────────────────────────────────

class _LoginBanner extends StatelessWidget {
  final VoidCallback onTap;
  const _LoginBanner({required this.onTap});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Padding(
      padding: const EdgeInsets.fromLTRB(20, 28, 20, 0),
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: cs.primaryContainer.withValues(alpha: 0.18),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(
              color: cs.primary.withValues(alpha: 0.35),
              width: 0.8,
            ),
            boxShadow: [
              BoxShadow(
                color: cs.primary.withValues(alpha: 0.12),
                blurRadius: 20,
                spreadRadius: -4,
                offset: const Offset(0, 4),
              ),
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.15),
                blurRadius: 10,
                offset: const Offset(0, 3),
              ),
            ],
          ),
          child: Row(
            children: [
              Icon(Icons.lock_open_rounded, color: cs.primary, size: 28),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Crée un compte gratuit',
                      style: tt.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                        color: cs.onSurface,
                      ),
                    ),
                    Text(
                      'Playlists, favoris, et bien plus',
                      style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                    ),
                  ],
                ),
              ),
              Icon(
                Icons.arrow_forward_ios_rounded,
                size: 16,
                color: cs.onSurfaceVariant,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Playlist card (horizontal carousel) ──────────────────────────────────────

class _PlaylistCard extends StatelessWidget {
  final PlaylistModel playlist;
  const _PlaylistCard({required this.playlist});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return GestureDetector(
      onTap: () => context.go('/library'),
      child: Container(
        decoration: BoxDecoration(
          color: cs.surfaceContainerHigh,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: cs.outlineVariant, width: 0.8),
          boxShadow: [
            BoxShadow(
              color: cs.primary.withValues(alpha: 0.07),
              blurRadius: 12,
              spreadRadius: -4,
            ),
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.12),
              blurRadius: 8,
              offset: const Offset(0, 3),
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(14, 14, 14, 12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Center(
                  child: Icon(
                    Icons.queue_music_rounded,
                    color: cs.primary.withValues(alpha: 0.65),
                    size: 36,
                  ),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                playlist.title,
                style: tt.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              Text(
                '${playlist.trackCount} titre${playlist.trackCount != 1 ? 's' : ''}',
                style: tt.labelSmall?.copyWith(color: cs.onSurfaceVariant),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Favorite chip (horizontal carousel) ──────────────────────────────────────

class _FavoriteChip extends StatelessWidget {
  final TrackModel track;
  const _FavoriteChip({required this.track});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Container(
      constraints: const BoxConstraints(maxWidth: 200),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: cs.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: cs.outlineVariant, width: 0.8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 8,
            offset: const Offset(0, 2),
            spreadRadius: -3,
          ),
        ],
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: cs.primaryContainer.withValues(alpha: 0.4),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(Icons.music_note_rounded, size: 15, color: cs.primary),
          ),
          const SizedBox(width: 10),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                track.title,
                style: tt.bodySmall?.copyWith(fontWeight: FontWeight.w600),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              if (track.artist.isNotEmpty)
                Text(
                  track.artist,
                  style: TextStyle(fontSize: 11, color: cs.onSurfaceVariant),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
            ],
          ),
          const SizedBox(width: 8),
          Icon(
            Icons.favorite_rounded,
            size: 12,
            color: cs.error.withValues(alpha: 0.65),
          ),
        ],
      ),
    );
  }
}

// ── Shared sub-widgets ────────────────────────────────────────────────────────

class _LiveBadge extends StatelessWidget {
  final bool large;
  const _LiveBadge({this.large = false});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: large ? 10 : 6,
        vertical: large ? 4 : 2,
      ),
      decoration: BoxDecoration(
        color: cs.error,
        borderRadius: BorderRadius.circular(6),
        boxShadow: [
          BoxShadow(
            color: cs.error.withValues(alpha: 0.4),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Text(
        'LIVE',
        style: TextStyle(
          color: cs.onError,
          fontSize: large ? 11 : 9,
          fontWeight: FontWeight.w800,
          letterSpacing: 0.5,
        ),
      ),
    );
  }
}

class _ListenerCount extends StatelessWidget {
  final int count;
  final bool light;
  const _ListenerCount({required this.count, this.light = false});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final color = light ? Colors.white70 : cs.onSurfaceVariant;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.headphones_rounded, size: 14, color: color),
        const SizedBox(width: 4),
        Text(
          '$count',
          style: TextStyle(
            fontSize: 12,
            color: color,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }
}

// Deterministic accent color per stream
Color _streamAccent(String id) {
  const palette = [
    Color(0xFF6366F1),
    Color(0xFF8B5CF6),
    Color(0xFFEC4899),
    Color(0xFF14B8A6),
    Color(0xFFF59E0B),
    Color(0xFF10B981),
    Color(0xFFEF4444),
    Color(0xFF3B82F6),
    Color(0xFF0EA5E9),
    Color(0xFFF97316),
  ];
  final hash = id.codeUnits.fold(0, (a, b) => a + b);
  return palette[hash % palette.length];
}
