import 'package:freezed_annotation/freezed_annotation.dart';
import '../../../domain/entities/stream.dart';

part 'audio_state.freezed.dart';

/// Current audio playback state
enum AudioPlaybackState { idle, loading, playing, paused, stopped, error }

/// Audio player state
@freezed
class AudioState with _$AudioState {
  const factory AudioState({
    required AudioPlaybackState playbackState,
    required double volume,
    required Duration position,
    required Duration duration,
    StreamEntity? currentStream,
    @Default('') String errorMessage,
    @Default(false) bool isShuffled,
    @Default(false) bool isLooping,
    @Default([]) List<StreamEntity> playlist,
    @Default(0) int playlistIndex,
  }) = _AudioState;

  const AudioState._();

  bool get isPlaying => playbackState == AudioPlaybackState.playing;
  bool get isPaused => playbackState == AudioPlaybackState.paused;
  bool get isLoading => playbackState == AudioPlaybackState.loading;
  bool get hasError => playbackState == AudioPlaybackState.error;

  /// Progress percentage (0.0 to 1.0)
  double get progress {
    if (duration.inMilliseconds == 0) return 0.0;
    return position.inMilliseconds / duration.inMilliseconds;
  }
}

/// Factory for initial audio state
class AudioStateFactory {
  static AudioState initial() => const AudioState(
        playbackState: AudioPlaybackState.idle,
        volume: 1.0,
        position: Duration.zero,
        duration: Duration.zero,
      );
}
