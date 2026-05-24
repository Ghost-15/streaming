import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../widgets/index.dart';
import '../../providers/stream_providers.dart';
import '../viewmodels/audio_view_model.dart';

/// Full-featured audio player screen with Material 3 design
class AudioPlayerScreen extends ConsumerWidget {
  final String? streamId;

  const AudioPlayerScreen({
    super.key,
    this.streamId,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final audioState = ref.watch(audioViewModelProvider);
    final currentStream = audioState.currentStream;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Audio Player'),
        elevation: 0,
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              // Now Playing Section
              if (audioState.isLoading)
                LoadingIndicator(
                  message: 'Loading stream...',
                )
              else if (audioState.hasError)
                _ErrorWidget(message: audioState.errorMessage)
              else if (currentStream != null)
                _NowPlayingSection(stream: currentStream)
              else
                _EmptyState(),

              const SizedBox(height: 32),

              // Playback Controls
              AudioControls(
                isPlaying: audioState.isPlaying,
                isLoading: audioState.isLoading,
                onPlay: () {
                  if (currentStream != null) {
                    if (audioState.isPaused) {
                      ref
                          .read(audioViewModelProvider.notifier)
                          .resume();
                    } else {
                      ref
                          .read(audioViewModelProvider.notifier)
                          .playStream(currentStream);
                    }
                  }
                },
                onPause: () {
                  ref.read(audioViewModelProvider.notifier).pause();
                },
                onStop: () {
                  ref.read(audioViewModelProvider.notifier).stop();
                },
              ),

              const SizedBox(height: 24),

              // Progress Bar
              if (audioState.duration != Duration.zero)
                _ProgressSection(
                  position: audioState.position,
                  duration: audioState.duration,
                  onSeek: (position) {
                    ref
                        .read(audioViewModelProvider.notifier)
                        .seek(position);
                  },
                ),

              const SizedBox(height: 24),

              // Volume Control
              VolumeControl(
                volume: audioState.volume,
                onVolumeChanged: (value) {
                  ref.read(audioViewModelProvider.notifier).setVolume(value);
                },
              ),

              const SizedBox(height: 24),

              // Playlist Mode Toggles
              _PlaylistModes(
                isShuffled: audioState.isShuffled,
                isLooping: audioState.isLooping,
                onToggleShuffle: () {
                  ref
                      .read(audioViewModelProvider.notifier)
                      .toggleShuffle();
                },
                onToggleLoop: () {
                  ref
                      .read(audioViewModelProvider.notifier)
                      .toggleLoop();
                },
              ),

              const SizedBox(height: 32),

              // Active Streams Section
              const _ActiveStreamsSection(),
            ],
          ),
        ),
      ),
    );
  }
}

/// Now Playing Section
class _NowPlayingSection extends StatelessWidget {
  final dynamic stream;

  const _NowPlayingSection({required this.stream});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Now Playing',
              style: Theme.of(context).textTheme.labelLarge?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 12),
            Text(
              stream.title,
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              'by ${stream.broadcasterName}',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
            ),
            if (stream.isLive) ...[
              const SizedBox(height: 12),
              Chip(
                label: const Text('LIVE'),
                backgroundColor: colorScheme.error,
                labelStyle: TextStyle(color: colorScheme.onError),
              ),
            ],
            const SizedBox(height: 12),
            Text(
              stream.description ?? 'No description',
              style: Theme.of(context).textTheme.bodySmall,
              maxLines: 3,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }
}

/// Empty State
class _EmptyState extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(48.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.music_note,
              size: 64,
              color: colorScheme.outline,
            ),
            const SizedBox(height: 16),
            Text(
              'No Stream Selected',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              'Select a stream from the list below',
              style: Theme.of(context).textTheme.bodySmall,
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

/// Error Widget
class _ErrorWidget extends StatelessWidget {
  final String? message;

  const _ErrorWidget({this.message});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      color: colorScheme.errorContainer,
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            Icon(
              Icons.error_outline,
              color: colorScheme.error,
              size: 32,
            ),
            const SizedBox(height: 8),
            Text(
              message ?? 'An error occurred',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: colorScheme.onErrorContainer,
                  ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

/// Progress Bar Section
class _ProgressSection extends StatelessWidget {
  final Duration position;
  final Duration duration;
  final ValueChanged<Duration> onSeek;

  const _ProgressSection({
    required this.position,
    required this.duration,
    required this.onSeek,
  });

  @override
  Widget build(BuildContext context) {
    final progress = duration.inMilliseconds > 0
        ? position.inMilliseconds / duration.inMilliseconds
        : 0.0;

    return Column(
      children: [
        Semantics(
          slider: true,
          label: 'Playback progress',
          onIncrease:
              progress < 1.0
                  ? () {
                    final newPosition = Duration(
                      milliseconds: position.inMilliseconds + 5000,
                    );
                    if (newPosition <= duration) {
                      onSeek(newPosition);
                    }
                  }
                  : null,
          onDecrease:
              progress > 0.0
                  ? () {
                    final newPosition = Duration(
                      milliseconds: (position.inMilliseconds - 5000).clamp(0, duration.inMilliseconds),
                    );
                    onSeek(newPosition);
                  }
                  : null,
          child: SliderTheme(
            data: SliderThemeData(
              trackHeight: 4,
              thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 8),
            ),
            child: Slider(
              value: progress.clamp(0.0, 1.0),
              onChanged: (value) {
                final newPosition = Duration(
                  milliseconds: (value * duration.inMilliseconds).toInt(),
                );
                onSeek(newPosition);
              },
            ),
          ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              _formatDuration(position),
              style: Theme.of(context).textTheme.labelSmall,
            ),
            Text(
              _formatDuration(duration),
              style: Theme.of(context).textTheme.labelSmall,
            ),
          ],
        ),
      ],
    );
  }

  String _formatDuration(Duration duration) {
    final minutes = duration.inMinutes;
    final seconds = duration.inSeconds % 60;
    return '$minutes:${seconds.toString().padLeft(2, '0')}';
  }
}

/// Playlist Mode Toggles
class _PlaylistModes extends StatelessWidget {
  final bool isShuffled;
  final bool isLooping;
  final VoidCallback onToggleShuffle;
  final VoidCallback onToggleLoop;

  const _PlaylistModes({
    required this.isShuffled,
    required this.isLooping,
    required this.onToggleShuffle,
    required this.onToggleLoop,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceEvenly,
      children: [
        Semantics(
          button: true,
          label: isShuffled ? 'Shuffle on' : 'Shuffle off',
          enabled: true,
          onTap: onToggleShuffle,
          child: FilledButton.tonal(
            onPressed: onToggleShuffle,
            style: FilledButton.styleFrom(
              backgroundColor:
                  isShuffled ? colorScheme.primary : colorScheme.surfaceVariant,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.shuffle,
                  color: isShuffled ? colorScheme.onPrimary : colorScheme.onSurfaceVariant,
                ),
                const SizedBox(width: 8),
                Text(isShuffled ? 'On' : 'Off'),
              ],
            ),
          ),
        ),
        Semantics(
          button: true,
          label: isLooping ? 'Loop on' : 'Loop off',
          enabled: true,
          onTap: onToggleLoop,
          child: FilledButton.tonal(
            onPressed: onToggleLoop,
            style: FilledButton.styleFrom(
              backgroundColor:
                  isLooping ? colorScheme.primary : colorScheme.surfaceVariant,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.repeat,
                  color: isLooping ? colorScheme.onPrimary : colorScheme.onSurfaceVariant,
                ),
                const SizedBox(width: 8),
                Text(isLooping ? 'On' : 'Off'),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

/// Active Streams Section
class _ActiveStreamsSection extends ConsumerWidget {
  const _ActiveStreamsSection();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final streamsAsync = ref.watch(activeStreamsProvider);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Available Streams',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: 16),
        streamsAsync.when(
          loading: () => const LoadingIndicator(message: 'Loading streams...'),
          error: (error, stackTrace) => _ErrorWidget(
            message: 'Failed to load streams',
          ),
          data: (streams) {
            if (streams.isEmpty) {
              return Card(
                child: Padding(
                  padding: const EdgeInsets.all(24.0),
                  child: Center(
                    child: Text(
                      'No streams available',
                      style: Theme.of(context).textTheme.bodyMedium,
                    ),
                  ),
                ),
              );
            }
            return ListView.separated(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: streams.length,
              separatorBuilder: (context, index) =>
                  const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final stream = streams[index];
                return StreamCard(
                  stream: stream,
                  onPlay: () {
                    ref
                        .read(audioViewModelProvider.notifier)
                        .playStream(stream);
                  },
                );
              },
            );
          },
        ),
      ],
    );
  }
}
