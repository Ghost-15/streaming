import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

// Non-Web fallback used by VM tests and unsupported desktop/mobile targets.
// The production Web export provides MediaSource and Media Session playback.
class AudioNotifier extends ChangeNotifier {
  final http.Client _streamClient;

  AudioNotifier({http.Client? streamClient, StreamRepository? repository})
    : _streamClient = streamClient ?? http.Client();

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1;
  StreamModel? _currentStream;
  String _errorMessage = '';

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;
  bool get isPlaying => _playbackState == AudioPlaybackState.playing;
  bool get isPaused => _playbackState == AudioPlaybackState.paused;
  bool get isLoading => _playbackState == AudioPlaybackState.loading;
  bool get isBuffering => _playbackState == AudioPlaybackState.buffering;
  bool get hasError => _playbackState == AudioPlaybackState.error;

  Future<void> playStream(StreamModel stream) async {
    _currentStream = stream;
    _errorMessage =
        'La lecture audio en direct est disponible dans le client Web.';
    _setState(AudioPlaybackState.error);
  }

  Future<void> pause() async {
    if (_currentStream != null) _setState(AudioPlaybackState.paused);
  }

  Future<void> resume() async {
    if (_currentStream != null && !hasError) {
      _setState(AudioPlaybackState.playing);
    }
  }

  Future<void> stop() async {
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0, 1);
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

  void _setState(AudioPlaybackState state) {
    _playbackState = state;
    notifyListeners();
  }

  @override
  void dispose() {
    _streamClient.close();
    super.dispose();
  }
}
