import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:just_audio/just_audio.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

/// Mobile/desktop (dart:io) live audio listener backed by just_audio.
///
/// The public surface mirrors the Web implementation (`audio_notifier_web.dart`)
/// exactly so the UI and providers stay platform-agnostic. just_audio handles
/// the network buffering of the live `GET /streams/:id/audio` response, which is
/// natively decodable by ExoPlayer (Android). Background playback, system
/// controls and audio-focus interruptions are wired in a later step.
class AudioNotifier extends ChangeNotifier {
  final StreamRepository _repository;
  final AudioPlayer _player = AudioPlayer();
  StreamSubscription<PlayerState>? _stateSub;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';
  String? _joinedStreamId;
  bool _userPaused = false;
  bool _disposed = false;

  AudioNotifier({StreamRepository? repository})
    : _repository = repository ?? const StreamRepository() {
    unawaited(_player.setVolume(_volume));
    _stateSub = _player.playerStateStream.listen(_onPlayerState);
  }

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;

  bool get isPlaying => _playbackState == AudioPlaybackState.playing;
  bool get isPaused => _playbackState == AudioPlaybackState.paused;
  bool get isLoading => _playbackState == AudioPlaybackState.loading;
  bool get isBuffering => _playbackState == AudioPlaybackState.buffering;
  bool get hasError => _playbackState == AudioPlaybackState.error;

  void _onPlayerState(PlayerState state) {
    if (_disposed || hasError) return;
    switch (state.processingState) {
      case ProcessingState.idle:
        // Ignore transient idle bursts emitted while (re)loading a source; our
        // own stop()/error paths own the idle and error transitions.
        return;
      case ProcessingState.loading:
        _setState(AudioPlaybackState.loading);
      case ProcessingState.buffering:
        _setState(
          _userPaused
              ? AudioPlaybackState.paused
              : AudioPlaybackState.buffering,
        );
      case ProcessingState.ready:
        _setState(
          state.playing
              ? AudioPlaybackState.playing
              : AudioPlaybackState.paused,
        );
      case ProcessingState.completed:
        _finish();
    }
  }

  Future<void> playStream(StreamModel stream) async {
    _userPaused = false;
    _currentStream = stream;
    _errorMessage = '';
    _setState(AudioPlaybackState.loading);

    unawaited(_joinStream(stream.id));
    try {
      await _player.setUrl(stream.streamUrl);
      await _player.play();
    } catch (e) {
      _fail('Impossible de lire le direct: $e');
    }
  }

  Future<void> _joinStream(String streamId) async {
    try {
      await _repository.joinStream(streamId);
      _joinedStreamId = streamId;
    } catch (_) {
      // The public audio endpoint stays usable for anonymous listeners; the
      // history/counting request is deliberately best-effort.
    }
  }

  void _leaveJoinedStream() {
    final streamId = _joinedStreamId;
    _joinedStreamId = null;
    if (streamId != null) {
      unawaited(_repository.leaveStream(streamId).catchError((_) {}));
    }
  }

  Future<void> pause() async {
    _userPaused = true;
    await _player.pause();
    _setState(AudioPlaybackState.paused);
  }

  Future<void> resume() async {
    _userPaused = false;
    final stream = _currentStream;
    if (stream == null) return;
    // A live stream must not resume from a stale buffered position: reload so
    // playback restarts at the current live edge.
    _setState(AudioPlaybackState.loading);
    try {
      await _player.setUrl(stream.streamUrl);
      await _player.play();
    } catch (e) {
      _fail('Lecture impossible: $e');
    }
  }

  Future<void> stop() async {
    _userPaused = false;
    await _player.stop();
    _leaveJoinedStream();
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0.0, 1.0);
    await _player.setVolume(_volume);
    notifyListeners();
  }

  Future<void> retry() async {
    final stream = _currentStream;
    if (stream != null) await playStream(stream);
  }

  void clearError() {
    _errorMessage = '';
    _setState(AudioPlaybackState.idle);
  }

  void _finish() {
    _leaveJoinedStream();
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  void _fail(String message) {
    _errorMessage = message;
    _leaveJoinedStream();
    _setState(AudioPlaybackState.error);
  }

  void _setState(AudioPlaybackState state) {
    if (_disposed) return;
    _playbackState = state;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    _leaveJoinedStream();
    unawaited(_stateSub?.cancel());
    unawaited(_player.dispose());
    super.dispose();
  }
}
