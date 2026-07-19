import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../api/models/stream_model.dart';
import '../notifiers/audio_notifier.dart';
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

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();
    final streams = context.watch<StreamNotifier>();

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => context.read<StreamNotifier>().loadActive(),
        child: CustomScrollView(
          slivers: [
            _AppBar(session: session),
            _GreetingSliver(
              greeting: _greeting(),
              session: session,
            ),
            if (streams.isLoading)
              const _LoadingSliver()
            else if (streams.streams.isEmpty)
              const SliverFillRemaining(child: _EmptyState())
            else ...[
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
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
              const _SectionHeader('Tous les streams'),
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
            if (!session.isAuthenticated)
              SliverToBoxAdapter(
                child: _LoginBanner(onTap: () => context.go('/register')),
              ),
            const SliverPadding(padding: EdgeInsets.only(bottom: 120)),
          ],
        ),
      ),
    );
  }
}

// ── App bar ───────────────────────────────────────────────────────────────────

class _AppBar extends StatelessWidget {
  final SessionNotifier session;
  const _AppBar({required this.session});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final user = session.user;

    return SliverAppBar(
      floating: true,
      snap: true,
      backgroundColor: cs.surfaceContainerLowest,
      surfaceTintColor: Colors.transparent,
      title: Row(
        children: [
          Container(
            width: 30,
            height: 30,
            decoration: BoxDecoration(
              color: cs.primary,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(Icons.radio_rounded, color: cs.onPrimary, size: 17),
          ),
          const SizedBox(width: 10),
          Text(
            'StreamPulse',
            style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w800, letterSpacing: -0.5),
          ),
        ],
      ),
      actions: [
        if (session.isAuthenticated && user != null) ...[
          if (user.role == Role.admin)
            IconButton(
              icon: const Icon(Icons.admin_panel_settings_outlined),
              tooltip: 'Administration',
              onPressed: () => context.go('/admin'),
            ),
          if (user.role == Role.diffuseur)
            IconButton(
              icon: const Icon(Icons.mic_none_rounded),
              tooltip: 'Diffuser',
              onPressed: () => context.go('/broadcaster'),
            ),
          Padding(
            padding: const EdgeInsets.only(right: 12),
            child: GestureDetector(
              onTap: () => _showAccountMenu(context, session),
              child: CircleAvatar(
                radius: 16,
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
          ),
        ] else
          Padding(
            padding: const EdgeInsets.only(right: 8),
            child: FilledButton.tonal(
              onPressed: () => context.go('/login'),
              style: FilledButton.styleFrom(
                minimumSize: const Size(0, 36),
                padding: const EdgeInsets.symmetric(horizontal: 16),
              ),
              child: const Text('Se connecter'),
            ),
          ),
      ],
    );
  }

  String _initial(String firstName, String email) {
    if (firstName.isNotEmpty) return firstName[0].toUpperCase();
    if (email.isNotEmpty) return email[0].toUpperCase();
    return '?';
  }

  void _showAccountMenu(BuildContext context, SessionNotifier session) {
    final cs = Theme.of(context).colorScheme;
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: cs.surfaceContainer,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 8),
            Container(
              width: 36,
              height: 4,
              decoration: BoxDecoration(
                color: cs.outlineVariant,
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.person_outline_rounded),
              title: Text(
                session.user!.email,
                style: const TextStyle(fontSize: 14),
              ),
              subtitle: Text(_roleLabel(session.user!.role)),
            ),
            const Divider(height: 1),
            ListTile(
              leading: const Icon(Icons.logout_rounded),
              title: const Text('Se déconnecter'),
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

// ── Greeting ──────────────────────────────────────────────────────────────────

class _GreetingSliver extends StatelessWidget {
  final String greeting;
  final SessionNotifier session;
  const _GreetingSliver({required this.greeting, required this.session});

  @override
  Widget build(BuildContext context) {
    final tt = Theme.of(context).textTheme;
    final cs = Theme.of(context).colorScheme;
    final name = session.user?.firstName;

    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 20, 20, 4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              name != null && name.isNotEmpty ? '$greeting, $name' : greeting,
              style: tt.headlineMedium?.copyWith(fontWeight: FontWeight.w800, letterSpacing: -0.5),
            ),
            const SizedBox(height: 4),
            Text(
              'Découvre les streams en direct',
              style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Section header ────────────────────────────────────────────────────────────

class _SectionHeader extends StatelessWidget {
  final String title;
  const _SectionHeader(this.title);

  @override
  Widget build(BuildContext context) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 28, 20, 14),
        child: Text(
          title,
          style: Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w700),
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
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
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
                    if (stream.isLive)
                      const _LiveBadge(large: true),
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
                        _ListenerCount(count: stream.listenerCount, light: true),
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
                          child: Icon(Icons.play_arrow_rounded, color: accent, size: 30),
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
              child: Icon(Icons.radio_outlined, size: 40, color: cs.onSurfaceVariant),
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
            color: cs.primaryContainer.withValues(alpha: 0.3),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: cs.primary.withValues(alpha: 0.3)),
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
              Icon(Icons.arrow_forward_ios_rounded, size: 16, color: cs.onSurfaceVariant),
            ],
          ),
        ),
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
        borderRadius: BorderRadius.circular(4),
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
