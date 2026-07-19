import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../notifiers/audio_notifier.dart';
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

    final progress = audio.duration.inMilliseconds > 0
        ? (audio.position.inMilliseconds / audio.duration.inMilliseconds)
            .clamp(0.0, 1.0)
        : 0.0;

    return GestureDetector(
      onTap: () => _openFullPlayer(context),
      child: Container(
        height: 68,
        decoration: BoxDecoration(
          color: cs.surface,
          border: Border(
            top: BorderSide(color: cs.outlineVariant, width: 0.5),
          ),
        ),
        child: Column(
          children: [
            // Thin progress line at very top
            LinearProgressIndicator(
              value: progress > 0 ? progress.toDouble() : null,
              backgroundColor: Colors.transparent,
              color: cs.primary,
              minHeight: 2,
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                child: Row(
                  children: [
                    // Artwork thumbnail
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Container(
                        width: 44,
                        height: 44,
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
                    // Title + broadcaster
                    Expanded(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            stream.title,
                            style: tt.bodyMedium
                                ?.copyWith(fontWeight: FontWeight.w600),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                          Text(
                            stream.broadcasterName,
                            style: tt.bodySmall
                                ?.copyWith(color: cs.onSurfaceVariant),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                    // Play / Pause
                    if (audio.isLoading || audio.isBuffering)
                      Padding(
                        padding: const EdgeInsets.all(12),
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
                        icon: Icon(
                          audio.isPlaying
                              ? Icons.pause_rounded
                              : Icons.play_arrow_rounded,
                          size: 28,
                        ),
                        color: cs.onSurface,
                        onPressed: () => audio.isPlaying
                            ? context.read<AudioNotifier>().pause()
                            : context.read<AudioNotifier>().resume(),
                      ),
                    // Stop / Close
                    IconButton(
                      icon: const Icon(Icons.close_rounded, size: 20),
                      color: cs.onSurfaceVariant,
                      onPressed: () => context.read<AudioNotifier>().stop(),
                    ),
                  ],
                ),
              ),
            ),
          ],
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

                  // Title + LIVE badge
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
                            const SizedBox(height: 4),
                            Text(
                              stream.broadcasterName,
                              style: tt.bodyMedium
                                  ?.copyWith(color: cs.onSurfaceVariant),
                            ),
                          ],
                        ),
                      ),
                      if (stream.isLive) ...[
                        const SizedBox(width: 12),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: cs.error,
                            borderRadius: BorderRadius.circular(6),
                          ),
                          child: Text(
                            'LIVE',
                            style: TextStyle(
                              color: cs.onError,
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),

                  // Listeners
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Icon(
                        Icons.headphones_rounded,
                        size: 14,
                        color: cs.onSurfaceVariant,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${stream.listenerCount} auditeur${stream.listenerCount != 1 ? 's' : ''}',
                        style: tt.bodySmall
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                    ],
                  ),
                  const SizedBox(height: 32),

                  // Progress bar (only if stream has duration)
                  if (audio.duration != Duration.zero) ...[
                    _ProgressBar(audio: audio),
                    const SizedBox(height: 24),
                  ],

                  // Main controls
                  AudioControls(
                    isPlaying: audio.isPlaying,
                    isLoading: audio.isLoading || audio.isBuffering,
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

                  // Shuffle / Loop
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _ModeChip(
                        icon: Icons.shuffle_rounded,
                        label: 'Aléatoire',
                        active: audio.isShuffled,
                        onTap: () =>
                            context.read<AudioNotifier>().toggleShuffle(),
                      ),
                      const SizedBox(width: 12),
                      _ModeChip(
                        icon: Icons.repeat_rounded,
                        label: 'Répéter',
                        active: audio.isLooping,
                        onTap: () =>
                            context.read<AudioNotifier>().toggleLoop(),
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

class _ProgressBar extends StatelessWidget {
  final AudioNotifier audio;
  const _ProgressBar({required this.audio});

  String _fmt(Duration d) {
    final m = d.inMinutes;
    final s = d.inSeconds % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final progress = audio.duration.inMilliseconds > 0
        ? (audio.position.inMilliseconds / audio.duration.inMilliseconds)
            .clamp(0.0, 1.0)
        : 0.0;

    return Column(
      children: [
        Slider(
          value: progress.toDouble(),
          onChanged: (v) => context.read<AudioNotifier>().seek(
                Duration(
                    milliseconds:
                        (v * audio.duration.inMilliseconds).toInt()),
              ),
        ),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(_fmt(audio.position),
                style: Theme.of(context).textTheme.labelSmall),
            Text(_fmt(audio.duration),
                style: Theme.of(context).textTheme.labelSmall),
          ],
        ),
      ],
    );
  }
}

class _ModeChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool active;
  final VoidCallback onTap;
  const _ModeChip({
    required this.icon,
    required this.label,
    required this.active,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        decoration: BoxDecoration(
          color: active
              ? cs.primaryContainer.withValues(alpha: 0.4)
              : cs.surfaceContainerHigh,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: active ? cs.primary : cs.outlineVariant,
            width: active ? 1.5 : 0.5,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 16,
              color: active ? cs.primary : cs.onSurfaceVariant,
            ),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight:
                    active ? FontWeight.w600 : FontWeight.w400,
                color: active ? cs.primary : cs.onSurfaceVariant,
              ),
            ),
          ],
        ),
      ),
    );
  }
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
