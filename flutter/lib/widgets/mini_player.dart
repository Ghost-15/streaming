import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../notifiers/audio_notifier.dart';
import '../notifiers/favorite_notifier.dart';
import '../notifiers/playlist_notifier.dart';
import 'audio_controls.dart';
import 'volume_control.dart';

// ── Mini player bar (persistent, above NavigationBar) ─────────────────────────

class MiniPlayer extends StatelessWidget {
  const MiniPlayer({super.key});

  @override
  Widget build(BuildContext context) {
    final audio = context.watch<AudioNotifier>();
    if (audio.currentStream == null) return const SizedBox.shrink();

    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final stream = audio.currentStream!;
    final accent = _streamAccent(stream.id);

    return Semantics(
      button: true,
      label: 'Ouvrir le lecteur : ${stream.title}',
      child: GestureDetector(
        onTap: () => _openFullPlayer(context),
        child: Container(
          margin: const EdgeInsets.fromLTRB(16, 0, 16, 8),
          decoration: BoxDecoration(
            color: cs.surfaceContainerHigh,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: cs.outlineVariant, width: 0.8),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.22),
                blurRadius: 18,
                offset: const Offset(0, 4),
              ),
              BoxShadow(
                color: accent.withValues(alpha: 0.06),
                blurRadius: 20,
                spreadRadius: -4,
              ),
            ],
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(22),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (audio.isLoading || audio.isBuffering)
                  LinearProgressIndicator(
                    backgroundColor: Colors.transparent,
                    color: accent,
                    minHeight: 2,
                  ),
                Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 8,
                  ),
                  child: Row(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(10),
                        child: Container(
                          width: 42,
                          height: 42,
                          decoration: BoxDecoration(
                            gradient: LinearGradient(
                              begin: Alignment.topLeft,
                              end: Alignment.bottomRight,
                              colors: [accent, accent.withValues(alpha: 0.5)],
                            ),
                          ),
                          child: const Icon(
                            Icons.radio_rounded,
                            color: Colors.white54,
                            size: 20,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              stream.title,
                              style: tt.bodyMedium?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            Text(
                              stream.broadcasterName,
                              style: tt.bodySmall?.copyWith(
                                color: cs.onSurfaceVariant,
                              ),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ],
                        ),
                      ),
                      if (audio.isLoading)
                        Padding(
                          padding: const EdgeInsets.all(10),
                          child: SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: cs.primary,
                            ),
                          ),
                        )
                      else
                        IconButton(
                          tooltip: audio.isPlaying || audio.isBuffering
                              ? 'Pause'
                              : 'Lecture',
                          icon: Icon(
                            audio.isPlaying || audio.isBuffering
                                ? Icons.pause_rounded
                                : Icons.play_arrow_rounded,
                            size: 26,
                          ),
                          color: cs.onSurface,
                          onPressed: () {
                            final notifier = context.read<AudioNotifier>();
                            if (audio.isPlaying || audio.isBuffering) {
                              notifier.pause();
                            } else if (audio.isPaused) {
                              notifier.resume();
                            } else {
                              notifier.playStream(stream);
                            }
                          },
                        ),
                      IconButton(
                        tooltip: 'Fermer le lecteur',
                        icon: const Icon(Icons.close_rounded, size: 18),
                        color: cs.onSurfaceVariant,
                        onPressed: () => context.read<AudioNotifier>().stop(),
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

  void _openFullPlayer(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      useSafeArea: false,
      backgroundColor: Colors.transparent,
      enableDrag: true,
      builder: (_) => ChangeNotifierProvider.value(
        value: context.read<AudioNotifier>(),
        child: const _FullPlayerSheet(),
      ),
    );
  }
}

// ── Full player sheet (slide-up modal) ────────────────────────────────────────

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
          borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
        ),
        child: Column(
          children: [
            // Drag handle
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
                  // Top bar: collapse button + label
                  Row(
                    children: [
                      IconButton(
                        icon: const Icon(
                          Icons.keyboard_arrow_down_rounded,
                          size: 28,
                        ),
                        tooltip: 'Réduire le lecteur',
                        onPressed: () => Navigator.of(context).pop(),
                        color: cs.onSurface,
                      ),
                      const Spacer(),
                      Text(
                        'En cours de lecture',
                        style: tt.labelLarge?.copyWith(
                          color: cs.onSurfaceVariant,
                        ),
                      ),
                      const Spacer(),
                      const SizedBox(width: 48),
                    ],
                  ),
                  const SizedBox(height: 20),

                  // Artwork
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
                            const Icon(
                              Icons.radio_rounded,
                              color: Colors.white38,
                              size: 72,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 28),
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              stream.title,
                              style: tt.headlineSmall?.copyWith(
                                fontWeight: FontWeight.w800,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            if (stream.isLive) ...[
                              const SizedBox(height: 4),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 7,
                                  vertical: 3,
                                ),
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
                              style: tt.bodyMedium?.copyWith(
                                color: cs.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(width: 12),
                      Row(
                        children: [
                          _ActionBtn(
                            label: isFavorited
                                ? 'Retirer des favoris'
                                : 'Ajouter aux favoris',
                            icon: isFavorited
                                ? Icons.favorite_rounded
                                : Icons.favorite_border_rounded,
                            color: isFavorited ? cs.error : cs.onSurfaceVariant,
                            onTap: () => isFavorited
                                ? context.read<FavoriteNotifier>().remove(
                                    stream.id,
                                  )
                                : context.read<FavoriteNotifier>().add(
                                    stream.id,
                                  ),
                          ),
                          _ActionBtn(
                            label: 'Ajouter à une playlist',
                            icon: Icons.playlist_add_rounded,
                            color: cs.onSurfaceVariant,
                            onTap: () =>
                                _showPlaylistPicker(context, stream.id),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),

                  // Main controls
                  AudioControls(
                    isPlaying: audio.isPlaying || audio.isBuffering,
                    isLoading: audio.isLoading,
                    onPlay: () => audio.isPaused
                        ? context.read<AudioNotifier>().resume()
                        : context.read<AudioNotifier>().playStream(stream),
                    onPause: () => context.read<AudioNotifier>().pause(),
                    onStop: () {
                      context.read<AudioNotifier>().stop();
                      Navigator.of(context).pop();
                    },
                  ),
                  const SizedBox(height: 28),

                  // Volume
                  VolumeControl(
                    volume: audio.volume,
                    onVolumeChanged: (v) =>
                        context.read<AudioNotifier>().setVolume(v),
                  ),
                  const SizedBox(height: 24),

                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.headphones_rounded,
                        size: 12,
                        color: cs.onSurfaceVariant,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${stream.listenerCount} auditeur${stream.listenerCount != 1 ? 's' : ''}',
                        style: tt.labelSmall?.copyWith(
                          color: cs.onSurfaceVariant,
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
    );
  }
}

// ── Sub-widgets ───────────────────────────────────────────────────────────────

class _ActionBtn extends StatelessWidget {
  final IconData icon;
  final Color color;
  final VoidCallback onTap;
  final String label;
  const _ActionBtn({
    required this.icon,
    required this.color,
    required this.onTap,
    required this.label,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      button: true,
      label: label,
      child: GestureDetector(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Icon(icon, color: color, size: 22),
        ),
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
          ? const Text(
              'Aucune playlist disponible.\nCrée-en une depuis ta bibliothèque.',
            )
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

// Déterministe: même couleur pour le même stream
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
