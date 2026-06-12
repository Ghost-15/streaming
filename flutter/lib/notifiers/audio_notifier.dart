import 'package:flutter/foundation.dart';
import 'package:just_audio/just_audio.dart';

import '../api/models/stream_model.dart';

enum AudioPlaybackState { idle, loading, playing, paused, stopped, error }

class AudioNotifier extends ChangeNotifier {
  final AudioPlayer _audioPlayer = AudioPlayer();

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  Duration _position = Duration.zero;
  Duration _duration = Duration.zero;
  StreamModel? _currentStream;
  String _errorMessage = '';
  bool _isShuffled = false;
  bool _isLooping = false;

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  Duration get position => _position;
  Duration get duration => _duration;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;
  bool get isShuffled => _isShuffled;
  bool get isLooping => _isLooping;

  bool get isPlaying => _playbackState == AudioPlaybackState.playing;
  bool get isPaused => _playbackState == AudioPlaybackState.paused;
  bool get isLoading => _playbackState == AudioPlaybackState.loading;
  bool get hasError => _playbackState == AudioPlaybackState.error;

  double get progress {
    if (_duration.inMilliseconds == 0) return 0.0;
    return _position.inMilliseconds / _duration.inMilliseconds;
  }

  AudioNotifier() {
    _initAudioPlayer();
  }

  void _initAudioPlayer() {
    _audioPlayer.playerStateStream.listen((state) {
      switch (state.processingState) {
        case ProcessingState.idle:
          _playbackState = AudioPlaybackState.idle;
          break;
        case ProcessingState.loading:
        case ProcessingState.buffering:
          _playbackState = AudioPlaybackState.loading;
          break;
        case ProcessingState.ready:
          _playbackState =
              state.playing ? AudioPlaybackState.playing : AudioPlaybackState.paused;
          break;
        case ProcessingState.completed:
          _playbackState = AudioPlaybackState.stopped;
          break;
      }
      notifyListeners();
    });

    _audioPlayer.positionStream.listen((pos) {
      _position = pos;
      notifyListeners();
    });

    _audioPlayer.durationStream.listen((dur) {
      _duration = dur ?? Duration.zero;
      notifyListeners();
    });

    _audioPlayer.volumeStream.listen((vol) {
      _volume = vol;
      notifyListeners();
    });
  }

  Future<void> playStream(StreamModel stream) async {
    try {
      _playbackState = AudioPlaybackState.loading;
      _currentStream = stream;
      notifyListeners();

      await _audioPlayer.setUrl(stream.streamUrl);
      await _audioPlayer.play();
    } catch (e) {
      _playbackState = AudioPlaybackState.error;
      _errorMessage = 'Failed to play stream: $e';
      notifyListeners();
    }
  }

  Future<void> pause() async {
    try {
      await _audioPlayer.pause();
    } catch (e) {
      _errorMessage = 'Failed to pause: $e';
      notifyListeners();
    }
  }

  Future<void> resume() async {
    try {
      await _audioPlayer.play();
    } catch (e) {
      _errorMessage = 'Failed to resume: $e';
      notifyListeners();
    }
  }

  Future<void> stop() async {
    try {
      await _audioPlayer.stop();
      _position = Duration.zero;
      notifyListeners();
    } catch (e) {
      _errorMessage = 'Failed to stop: $e';
      notifyListeners();
    }
  }

  Future<void> setVolume(double volume) async {
    await _audioPlayer.setVolume(volume.clamp(0.0, 1.0));
  }

  Future<void> seek(Duration position) async {
    await _audioPlayer.seek(position);
  }

  void toggleShuffle() {
    _isShuffled = !_isShuffled;
    notifyListeners();
  }

  void toggleLoop() {
    _isLooping = !_isLooping;
    notifyListeners();
  }

  void clearError() {
    _errorMessage = '';
    notifyListeners();
  }

  @override
  void dispose() {
    _audioPlayer.dispose();
    super.dispose();
  }
}
