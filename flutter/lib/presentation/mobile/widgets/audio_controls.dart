import 'package:flutter/material.dart';

/// Audio playback controls widget
/// Provides play, pause, stop buttons with proper accessibility
class AudioControls extends StatelessWidget {
  final bool isPlaying;
  final bool isLoading;
  final VoidCallback onPlay;
  final VoidCallback onPause;
  final VoidCallback onStop;

  const AudioControls({
    super.key,
    required this.isPlaying,
    required this.isLoading,
    required this.onPlay,
    required this.onPause,
    required this.onStop,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Semantics(
      container: true,
      label: 'Audio playback controls',
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Stop button
          Semantics(
            button: true,
            enabled: !isLoading,
            onTap: isLoading ? null : onStop,
            label: 'Stop playback',
            child: SizedBox(
              width: 56,
              height: 56,
              child: Material(
                color: colorScheme.primaryContainer,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(28),
                ),
                child: InkWell(
                  onTap: isLoading ? null : onStop,
                  customBorder: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(28),
                  ),
                  child: Icon(
                    Icons.stop,
                    color: colorScheme.onPrimaryContainer,
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 16),

          // Play/Pause toggle button (main action)
          Semantics(
            button: true,
            enabled: !isLoading,
            onTap: isLoading ? null : (isPlaying ? onPause : onPlay),
            label: isPlaying ? 'Pause playback' : 'Play audio',
            child: SizedBox(
              width: 72,
              height: 72,
              child: Material(
                color: colorScheme.primary,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(36),
                ),
                elevation: 4,
                child: InkWell(
                  onTap: isLoading ? null : (isPlaying ? onPause : onPlay),
                  customBorder: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(36),
                  ),
                  child: isLoading
                      ? SizedBox(
                          width: 32,
                          height: 32,
                          child: CircularProgressIndicator(
                            valueColor: AlwaysStoppedAnimation<Color>(
                              colorScheme.onPrimary,
                            ),
                            strokeWidth: 2,
                          ),
                        )
                      : Icon(
                          isPlaying ? Icons.pause : Icons.play_arrow,
                          color: colorScheme.onPrimary,
                          size: 36,
                        ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 16),

          // Skip button (placeholder)
          Semantics(
            button: true,
            enabled: !isLoading,
            onTap: isLoading ? null : () {},
            label: 'Skip next',
            child: SizedBox(
              width: 56,
              height: 56,
              child: Material(
                color: colorScheme.primaryContainer,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(28),
                ),
                child: InkWell(
                  onTap: isLoading ? null : () {},
                  customBorder: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(28),
                  ),
                  child: Icon(
                    Icons.skip_next,
                    color: colorScheme.onPrimaryContainer,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
