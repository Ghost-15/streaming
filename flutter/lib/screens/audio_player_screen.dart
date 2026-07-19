import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/stream_model.dart';
import '../notifiers/audio_notifier.dart';
import '../notifiers/favorite_notifier.dart';
import '../notifiers/playlist_notifier.dart';
import '../notifiers/stream_notifier.dart';
import '../widgets/page_header.dart';
import '../widgets/stream_card.dart';

class AudioPlayerScreen extends StatefulWidget {
  final String? streamId;
  const AudioPlayerScreen({super.key, this.streamId});

  @override
  State<AudioPlayerScreen> createState() => _AudioPlayerScreenState();
}

class _AudioPlayerScreenState extends State<AudioPlayerScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<StreamNotifier>().loadActive();
    });
  }

  @override
  Widget build(BuildContext context) {
    final streams = context.watch<StreamNotifier>();
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => context.read<StreamNotifier>().loadActive(),
        child: CustomScrollView(
          slivers: [
            // ── Header ────────────────────────────────────────────────────
            SliverToBoxAdapter(
              child: PageHeader(
                icon: Icons.radio_rounded,
                title: 'Live',
                subtitle: 'Streams en direct',
                actions: [
                  IconButton(
                    icon: const Icon(Icons.refresh_rounded, size: 20),
                    color: cs.onSurfaceVariant,
                    onPressed: () => context.read<StreamNotifier>().loadActive(),
                  ),
                ],
              ),
            ),

            // ── Section title ─────────────────────────────────────────────
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 24, 20, 12),
                child: Text(
                  'Streams disponibles',
                  style: tt.titleMedium?.copyWith(fontWeight: FontWeight.w700),
                ),
              ),
            ),

            // ── Stream list ───────────────────────────────────────────────
            if (streams.isLoading)
              const SliverFillRemaining(
                child: Center(
                  child: CircularProgressIndicator(),
                ),
              )
            else if (streams.error != null && streams.streams.isEmpty)
              SliverFillRemaining(
                child: _ErrorState(
                  message: streams.error!,
                  onRetry: () => context.read<StreamNotifier>().loadActive(),
                ),
              )
            else if (streams.streams.isEmpty)
              const SliverFillRemaining(child: _EmptyState())
            else
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 120),
                sliver: SliverList.separated(
                  itemCount: streams.streams.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 12),
                  itemBuilder: (_, i) => StreamCard(
                    stream: streams.streams[i],
                    onPlay: () => context
                        .read<AudioNotifier>()
                        .playStream(streams.streams[i]),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

// ── Reuse full player from mini_player.dart via bottom sheet ─────────────────

// ignore: unused_element — accessed via showModalBottomSheet in parent
class _FullPlayerSheet extends StatelessWidget {
  const _FullPlayerSheet();

  @override
  Widget build(BuildContext context) {
    final audio = context.watch<AudioNotifier>();
    final stream = audio.currentStream;
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    if (stream == null) {
      WidgetsBinding.instance.addPostFrameCallback(
        (_) => Navigator.of(context).pop(),
      );
      return const SizedBox.shrink();
    }

    final accent = _streamAccent(stream.id);
    final favorites = context.watch<FavoriteNotifier>();
    final isFavorited = favorites.tracks.any((t) => t.id == stream.id);

    return DraggableScrollableSheet(
      initialChildSize: 0.95,
      minChildSize: 0.5,
      maxChildSize: 0.95,
      snap: true,
      snapSizes: const [0.5, 0.95],
      builder: (_, scrollController) => Container(
        decoration: BoxDecoration(
          color: cs.surfaceContainerLowest,
          borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
        ),
        child: Column(
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
            Expanded(
              child: ListView(
                controller: scrollController,
                padding: const EdgeInsets.fromLTRB(24, 0, 24, 48),
                children: [
                  Row(
                    children: [
                      IconButton(
                        icon: const Icon(
                          Icons.keyboard_arrow_down_rounded,
                          size: 28,
                        ),
                        onPressed: () => Navigator.of(context).pop(),
                        color: cs.onSurface,
                      ),
                      const Spacer(),
                      Text(
                        'En cours de lecture',
                        style: tt.labelLarge
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                      const Spacer(),
                      const SizedBox(width: 48),
                    ],
                  ),
                  const SizedBox(height: 20),
                  Center(
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(20),
                      child: Container(
                        width: 260,
                        height: 260,
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: [
                              accent,
                              accent.withValues(alpha: 0.4),
                              cs.surfaceContainerHigh,
                            ],
                          ),
                        ),
                        child: Stack(
                          alignment: Alignment.center,
                          children: [
                            Icon(
                              Icons.radio_rounded,
                              size: 180,
                              color: Colors.white.withValues(alpha: 0.06),
                            ),
                            const Icon(Icons.radio_rounded,
                                color: Colors.white38, size: 72),
                          ],
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              stream.title,
                              style: tt.headlineSmall
                                  ?.copyWith(fontWeight: FontWeight.w800),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            if (stream.isLive) ...[
                              const SizedBox(height: 4),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 7, vertical: 3),
                                decoration: BoxDecoration(
                                  color: cs.error,
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Text(
                                  'LIVE',
                                  style: TextStyle(
                                    color: cs.onError,
                                    fontSize: 10,
                                    fontWeight: FontWeight.w800,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                              ),
                            ],
                            const SizedBox(height: 4),
                            Text(
                              stream.broadcasterName,
                              style: tt.bodyMedium
                                  ?.copyWith(color: cs.onSurfaceVariant),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Row(
                        children: [
                          _ActionBtn(
                            icon: isFavorited
                                ? Icons.favorite_rounded
                                : Icons.favorite_border_rounded,
                            color: isFavorited
                                ? cs.error
                                : cs.onSurfaceVariant,
                            onTap: () => isFavorited
                                ? context
                                    .read<FavoriteNotifier>()
                                    .remove(stream.id)
                                : context
                                    .read<FavoriteNotifier>()
                                    .add(stream.id),
                          ),
                          _ActionBtn(
                            icon: Icons.playlist_add_rounded,
                            color: cs.onSurfaceVariant,
                            onTap: () =>
                                _showPlaylistPicker(context, stream.id),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 28),
                  _PlayerControls(audio: audio, stream: stream),
                  const SizedBox(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.headphones_rounded,
                          size: 12, color: cs.onSurfaceVariant),
                      const SizedBox(width: 4),
                      Text(
                        '${stream.listenerCount} auditeur${stream.listenerCount != 1 ? 's' : ''}',
                        style: tt.labelSmall
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                    ],
                  ),
                  const SizedBox(height: 32),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PlayerControls extends StatelessWidget {
  final AudioNotifier audio;
  final StreamModel stream;
  const _PlayerControls({required this.audio, required this.stream});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        // Stop
        _ControlBtn(
          icon: Icons.stop_rounded,
          size: 52,
          iconSize: 26,
          color: cs.surfaceContainerHighest,
          iconColor: cs.onSurface,
          onTap: () {
            context.read<AudioNotifier>().stop();
            Navigator.of(context).pop();
          },
        ),
        const SizedBox(width: 20),
        // Play / Pause main
        if (audio.isLoading || audio.isBuffering)
          Container(
            width: 68,
            height: 68,
            decoration: BoxDecoration(
              color: cs.primary,
              shape: BoxShape.circle,
            ),
            child: const Padding(
              padding: EdgeInsets.all(18),
              child: CircularProgressIndicator(
                  strokeWidth: 2.5, color: Colors.white),
            ),
          )
        else
          _ControlBtn(
            icon: audio.isPlaying
                ? Icons.pause_rounded
                : Icons.play_arrow_rounded,
            size: 68,
            iconSize: 34,
            color: cs.primary,
            iconColor: cs.onPrimary,
            onTap: () => audio.isPlaying
                ? context.read<AudioNotifier>().pause()
                : audio.isPaused
                    ? context.read<AudioNotifier>().resume()
                    : context.read<AudioNotifier>().playStream(stream),
          ),
        const SizedBox(width: 20),
        // Volume mute toggle
        _ControlBtn(
          icon: audio.volume > 0
              ? Icons.volume_up_rounded
              : Icons.volume_off_rounded,
          size: 52,
          iconSize: 26,
          color: cs.surfaceContainerHighest,
          iconColor: cs.onSurface,
          onTap: () => context
              .read<AudioNotifier>()
              .setVolume(audio.volume > 0 ? 0 : 1),
        ),
      ],
    );
  }
}

class _ControlBtn extends StatelessWidget {
  final IconData icon;
  final double size;
  final double iconSize;
  final Color color;
  final Color iconColor;
  final VoidCallback onTap;

  const _ControlBtn({
    required this.icon,
    required this.size,
    required this.iconSize,
    required this.color,
    required this.iconColor,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        child: Icon(icon, color: iconColor, size: iconSize),
      ),
    );
  }
}

class _ActionBtn extends StatelessWidget {
  final IconData icon;
  final Color color;
  final VoidCallback onTap;
  const _ActionBtn({
    required this.icon,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Icon(icon, color: color, size: 22),
      ),
    );
  }
}

void _showPlaylistPicker(BuildContext context, String trackId) {
  final playlists = context.read<PlaylistNotifier>();
  showDialog<void>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('Ajouter à une playlist'),
      content: playlists.playlists.isEmpty
          ? const Text('Aucune playlist disponible.\nCrée-en une depuis ta bibliothèque.')
          : SizedBox(
              width: double.maxFinite,
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: playlists.playlists.length,
                itemBuilder: (_, i) {
                  final pl = playlists.playlists[i];
                  return ListTile(
                    leading: const Icon(Icons.queue_music_rounded),
                    title: Text(pl.title),
                    subtitle: Text(
                      '${pl.trackCount} titre${pl.trackCount != 1 ? 's' : ''}',
                    ),
                    onTap: () {
                      context.read<PlaylistNotifier>().addTrack(pl.id, trackId);
                      Navigator.pop(ctx);
                    },
                  );
                },
              ),
            ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(ctx),
          child: const Text('Annuler'),
        ),
      ],
    ),
  );
}

// ── Empty / Error states ──────────────────────────────────────────────────────

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
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.radio_outlined,
                  size: 36, color: cs.onSurfaceVariant),
            ),
            const SizedBox(height: 20),
            Text(
              'Aucun stream en direct',
              style: tt.titleMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 8),
            Text(
              'Reviens plus tard ou invite un diffuseur à lancer un live.',
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  const _ErrorState({required this.message, required this.onRetry});

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
            Icon(Icons.wifi_off_rounded, size: 48, color: cs.onSurfaceVariant),
            const SizedBox(height: 16),
            Text(
              'Impossible de charger les streams',
              style: tt.titleSmall?.copyWith(fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh_rounded, size: 18),
              label: const Text('Réessayer'),
            ),
          ],
        ),
      ),
    );
  }
}

Color _streamAccent(String id) {
  const palette = [
    Color(0xFF6366F1), Color(0xFF8B5CF6), Color(0xFFEC4899),
    Color(0xFF14B8A6), Color(0xFFF59E0B), Color(0xFF10B981),
    Color(0xFFEF4444), Color(0xFF3B82F6), Color(0xFF0EA5E9),
    Color(0xFFF97316),
  ];
  final hash = id.codeUnits.fold(0, (a, b) => a + b);
  return palette[hash % palette.length];
}
