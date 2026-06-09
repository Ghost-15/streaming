import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/stream_model.dart';
import '../notifiers/audio_notifier.dart';
import '../notifiers/stream_notifier.dart';
import '../widgets/audio_controls.dart';
import '../widgets/loading_indicator.dart';
import '../widgets/stream_card.dart';
import '../widgets/volume_control.dart';

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
    final audio = context.watch<AudioNotifier>();
    final streams = context.watch<StreamNotifier>();
    final currentStream = audio.currentStream;

    return Scaffold(
      appBar: AppBar(title: const Text('Audio Player'), elevation: 0),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            children: [
              if (audio.isLoading)
                const LoadingIndicator(message: 'Loading stream...')
              else if (audio.hasError)
                _ErrorCard(message: audio.errorMessage)
              else if (currentStream != null)
                _NowPlayingCard(stream: currentStream)
              else
                const _EmptyStateCard(),

              const SizedBox(height: 32),

              AudioControls(
                isPlaying: audio.isPlaying,
                isLoading: audio.isLoading,
                onPlay: () {
                  if (currentStream != null) {
                    audio.isPaused
                        ? context.read<AudioNotifier>().resume()
                        : context.read<AudioNotifier>().playStream(currentStream);
                  }
                },
                onPause: () => context.read<AudioNotifier>().pause(),
                onStop: () => context.read<AudioNotifier>().stop(),
              ),

              const SizedBox(height: 24),

              if (audio.duration != Duration.zero)
                _ProgressSection(
                  position: audio.position,
                  duration: audio.duration,
                  onSeek: (pos) => context.read<AudioNotifier>().seek(pos),
                ),

              const SizedBox(height: 24),

              VolumeControl(
                volume: audio.volume,
                onVolumeChanged: (v) => context.read<AudioNotifier>().setVolume(v),
              ),

              const SizedBox(height: 24),

              _PlaylistModes(
                isShuffled: audio.isShuffled,
                isLooping: audio.isLooping,
                onToggleShuffle: () => context.read<AudioNotifier>().toggleShuffle(),
                onToggleLoop: () => context.read<AudioNotifier>().toggleLoop(),
              ),

              const SizedBox(height: 32),

              _ActiveStreamsSection(
                streams: streams.streams,
                isLoading: streams.isLoading,
                error: streams.error,
                onPlay: (stream) => context.read<AudioNotifier>().playStream(stream),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _NowPlayingCard extends StatelessWidget {
  final StreamModel stream;
  const _NowPlayingCard({required this.stream});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
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
            Text(stream.title, style: Theme.of(context).textTheme.headlineSmall),
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
            if (stream.description.isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(
                stream.description,
                style: Theme.of(context).textTheme.bodySmall,
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _EmptyStateCard extends StatelessWidget {
  const _EmptyStateCard();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(48),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.music_note, size: 64, color: colorScheme.outline),
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

class _ErrorCard extends StatelessWidget {
  final String? message;
  const _ErrorCard({this.message});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      color: colorScheme.errorContainer,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Icon(Icons.error_outline, color: colorScheme.error, size: 32),
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

class _ProgressSection extends StatelessWidget {
  final Duration position;
  final Duration duration;
  final ValueChanged<Duration> onSeek;

  const _ProgressSection({
    required this.position,
    required this.duration,
    required this.onSeek,
  });

  String _fmt(Duration d) {
    final m = d.inMinutes;
    final s = d.inSeconds % 60;
    return '$m:${s.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final progress = duration.inMilliseconds > 0
        ? (position.inMilliseconds / duration.inMilliseconds).clamp(0.0, 1.0)
        : 0.0;

    return Column(
      children: [
        Semantics(
          slider: true,
          label: 'Playback progress',
          child: SliderTheme(
            data: const SliderThemeData(
              trackHeight: 4,
              thumbShape: RoundSliderThumbShape(enabledThumbRadius: 8),
            ),
            child: Slider(
              value: progress,
              onChanged: (v) => onSeek(
                Duration(milliseconds: (v * duration.inMilliseconds).toInt()),
              ),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(_fmt(position), style: Theme.of(context).textTheme.labelSmall),
            Text(_fmt(duration), style: Theme.of(context).textTheme.labelSmall),
          ],
        ),
      ],
    );
  }
}

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
          child: FilledButton.tonal(
            onPressed: onToggleShuffle,
            style: FilledButton.styleFrom(
              backgroundColor: isShuffled ? colorScheme.primary : colorScheme.surfaceContainerHighest,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.shuffle,
                    color: isShuffled ? colorScheme.onPrimary : colorScheme.onSurfaceVariant),
                const SizedBox(width: 8),
                Text(isShuffled ? 'On' : 'Off'),
              ],
            ),
          ),
        ),
        Semantics(
          button: true,
          label: isLooping ? 'Loop on' : 'Loop off',
          child: FilledButton.tonal(
            onPressed: onToggleLoop,
            style: FilledButton.styleFrom(
              backgroundColor: isLooping ? colorScheme.primary : colorScheme.surfaceContainerHighest,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.repeat,
                    color: isLooping ? colorScheme.onPrimary : colorScheme.onSurfaceVariant),
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

class _ActiveStreamsSection extends StatelessWidget {
  final List<StreamModel> streams;
  final bool isLoading;
  final String? error;
  final void Function(StreamModel) onPlay;

  const _ActiveStreamsSection({
    required this.streams,
    required this.isLoading,
    required this.error,
    required this.onPlay,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Available Streams', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 16),
        if (isLoading)
          const LoadingIndicator(message: 'Loading streams...')
        else if (error != null)
          const _ErrorCard(message: 'Failed to load streams')
        else if (streams.isEmpty)
          Card(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Center(
                child: Text('No streams available',
                    style: Theme.of(context).textTheme.bodyMedium),
              ),
            ),
          )
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: streams.length,
            separatorBuilder: (_, _) => const SizedBox(height: 12),
            itemBuilder: (_, i) => StreamCard(
              stream: streams[i],
              onPlay: () => onPlay(streams[i]),
            ),
          ),
      ],
    );
  }
}
