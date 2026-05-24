import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:just_audio/just_audio.dart';
import '../../../domain/entities/stream.dart';
import '../../../data/repositories/stream_repository_impl.dart';
import 'audio_state.dart';

/// AudioViewModel StateNotifier for managing audio playback
class AudioViewModel extends StateNotifier<AudioState> {
  final AudioPlayer _audioPlayer;
  final StreamRepositoryImpl _streamRepository;

  AudioViewModel({
    required AudioPlayer audioPlayer,
    required StreamRepositoryImpl streamRepository,
  })  : _audioPlayer = audioPlayer,
        _streamRepository = streamRepository,
        super(AudioStateFactory.initial()) {
    _initializeAudioPlayer();
  }

  /// Initialize audio player listeners
  void _initializeAudioPlayer() {
    // Listen to playback state changes
    _audioPlayer.playerStateStream.listen((playerState) {
      final newPlaybackState = _mapPlayerStateToAudioState(playerState);
      state = state.copyWith(playbackState: newPlaybackState);
    });

    // Listen to position changes
    _audioPlayer.positionStream.listen((position) {
      state = state.copyWith(position: position);
    });

    // Listen to duration changes
    _audioPlayer.durationStream.listen((duration) {
      state = state.copyWith(duration: duration ?? Duration.zero);
    });

    // Listen to volume changes
    _audioPlayer.volumeStream.listen((volume) {
      state = state.copyWith(volume: volume);
    });
  }

  /// Play audio from stream
  Future<void> playStream(StreamEntity stream) async {
    try {
      state = state.copyWith(
        playbackState: AudioPlaybackState.loading,
        currentStream: stream,
      );

      // Set audio source to stream URL
      // TODO: Update with actual stream URL from StreamEntity
      await _audioPlayer.setUrl(stream.id);
      await _audioPlayer.play();

      state = state.copyWith(playbackState: AudioPlaybackState.playing);
    } catch (e) {
      state = state.copyWith(
        playbackState: AudioPlaybackState.error,
        errorMessage: 'Failed to play stream: $e',
      );
    }
  }

  /// Pause playback
  Future<void> pause() async {
    try {
      await _audioPlayer.pause();
      state = state.copyWith(playbackState: AudioPlaybackState.paused);
    } catch (e) {
      state = state.copyWith(
        playbackState: AudioPlaybackState.error,
        errorMessage: 'Failed to pause: $e',
      );
    }
  }

  /// Resume playback
  Future<void> resume() async {
    try {
      await _audioPlayer.play();
      state = state.copyWith(playbackState: AudioPlaybackState.playing);
    } catch (e) {
      state = state.copyWith(
        playbackState: AudioPlaybackState.error,
        errorMessage: 'Failed to resume: $e',
      );
    }
  }

  /// Stop playback
  Future<void> stop() async {
    try {
      await _audioPlayer.stop();
      state = state.copyWith(
        playbackState: AudioPlaybackState.stopped,
        position: Duration.zero,
      );
    } catch (e) {
      state = state.copyWith(
        playbackState: AudioPlaybackState.error,
        errorMessage: 'Failed to stop: $e',
      );
    }
  }

  /// Set volume (0.0 to 1.0)
  Future<void> setVolume(double volume) async {
    try {
      final clampedVolume = volume.clamp(0.0, 1.0);
      await _audioPlayer.setVolume(clampedVolume);
      state = state.copyWith(volume: clampedVolume);
    } catch (e) {
      state = state.copyWith(errorMessage: 'Failed to set volume: $e');
    }
  }

  /// Seek to position
  Future<void> seek(Duration position) async {
    try {
      await _audioPlayer.seek(position);
      state = state.copyWith(position: position);
    } catch (e) {
      state = state.copyWith(errorMessage: 'Failed to seek: $e');
    }
  }

  /// Toggle shuffle
  void toggleShuffle() {
    state = state.copyWith(isShuffled: !state.isShuffled);
  }

  /// Toggle loop mode
  void toggleLoop() {
    state = state.copyWith(isLooping: !state.isLooping);
  }

  /// Map AudioPlayer state to AudioPlaybackState
  AudioPlaybackState _mapPlayerStateToAudioState(PlayerState playerState) {
    switch (playerState.processingState) {
      case ProcessingState.idle:
        return AudioPlaybackState.idle;
      case ProcessingState.loading:
        return AudioPlaybackState.loading;
      case ProcessingState.buffering:
        return AudioPlaybackState.loading;
      case ProcessingState.ready:
        return playerState.playing ? AudioPlaybackState.playing : AudioPlaybackState.paused;
      case ProcessingState.completed:
        return AudioPlaybackState.stopped;
    }
  }

  /// Clear error message
  void clearError() {
    state = state.copyWith(errorMessage: '');
  }

  @override
  void dispose() {
    _audioPlayer.dispose();
    super.dispose();
  }
}

/// Provider for AudioPlayer singleton
final audioPlayerProvider = Provider<AudioPlayer>((ref) {
  return AudioPlayer();
});

/// Provider for StreamRepository
final streamRepositoryProvider = Provider<StreamRepositoryImpl>((ref) {
  return StreamRepositoryImpl();
});

/// StateNotifierProvider for AudioViewModel
final audioViewModelProvider =
    StateNotifierProvider<AudioViewModel, AudioState>((ref) {
  final audioPlayer = ref.watch(audioPlayerProvider);
  final streamRepository = ref.watch(streamRepositoryProvider);

  return AudioViewModel(
    audioPlayer: audioPlayer,
    streamRepository: streamRepository,
  );
});
