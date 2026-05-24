import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../viewmodels/audio_view_model.dart';

/// Home screen - audio player interface
class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final audioState = ref.watch(audioViewModelProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('StreamPulse'),
        centerTitle: true,
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // App title
              Text(
                'StreamPulse',
                style: Theme.of(context).textTheme.displaySmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),

              // Subtitle
              Text(
                'Live Audio Streaming',
                style: Theme.of(context).textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),

              // Playback state display
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      Text(
                        'Playback State',
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      const SizedBox(height: 12),
                      Text(
                        audioState.playbackState.toString(),
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                      const SizedBox(height: 12),
                      // Volume slider
                      Slider(
                        value: audioState.volume,
                        min: 0,
                        max: 1,
                        onChanged: (value) {
                          ref.read(audioViewModelProvider.notifier).setVolume(value);
                        },
                      ),
                      Text(
                        'Volume: ${(audioState.volume * 100).toStringAsFixed(0)}%',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 32),

              // Theme toggle button
              ElevatedButton(
                onPressed: () {
                  // Theme toggle will be implemented in app.dart with brightness provider
                },
                child: const Text('Toggle Dark Mode'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
