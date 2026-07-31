import 'dart:js_interop';

import 'package:flutter/foundation.dart';
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState {
  idle,
  loading,
  buffering,
  playing,
  paused,
  stopped,
  error,
}

class AudioNotifier extends ChangeNotifier {
  web.HTMLAudioElement? _el;
  int _version = 0;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  Duration get position => Duration.zero;
  Duration get duration => Duration.zero;
  double get progress => 0.0;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;
  bool get isShuffled => false;
  bool get isLooping => false;

  bool get isPlaying => _playbackState == AudioPlaybackState.playing;
  bool get isPaused => _playbackState == AudioPlaybackState.paused;
  bool get isLoading => _playbackState == AudioPlaybackState.loading;
  bool get isBuffering => _playbackState == AudioPlaybackState.buffering;
  bool get hasError => _playbackState == AudioPlaybackState.error;

  Future<void> playStream(StreamModel stream) async {
    final v = ++_version;
    _stopElement();

    _playbackState = AudioPlaybackState.loading;
    _currentStream = stream;
    notifyListeners();

    const StreamRepository().joinStream(stream.id).catchError((_) {});

    final el = web.HTMLAudioElement();
    _el = el;
    el.volume = _volume;
    // Required for CORS: ensures the browser sends Origin header so the server
    // can add Access-Control-Allow-Origin and the response is readable.
    el.crossOrigin = 'anonymous';

    el.addEventListener(
      'waiting',
      ((web.Event _) {
        debugPrint('[Audio] waiting');
        if (_version == v) _setState(AudioPlaybackState.buffering);
      }).toJS,
    );
    el.addEventListener(
      'stalled',
      ((web.Event _) {
        debugPrint('[Audio] stalled');
        if (_version == v) _setState(AudioPlaybackState.buffering);
      }).toJS,
    );
    el.addEventListener(
      'playing',
      ((web.Event _) {
        debugPrint('[Audio] playing');
        if (_version == v) _setState(AudioPlaybackState.playing);
      }).toJS,
    );
    el.addEventListener(
      'pause',
      ((web.Event _) {
        debugPrint('[Audio] pause');
        if (_version == v && _playbackState != AudioPlaybackState.idle) {
          _setState(AudioPlaybackState.paused);
        }
      }).toJS,
    );
    el.addEventListener(
      'ended',
      ((web.Event _) {
        debugPrint('[Audio] ended — broadcaster stopped');
        if (_version == v) {
          _stopElement();
          _currentStream = null;
          _setState(AudioPlaybackState.idle);
        }
      }).toJS,
    );
    el.addEventListener(
      'error',
      ((web.Event _) {
        debugPrint('[Audio] error');
        if (_version == v) {
          _errorMessage = 'Erreur de lecture audio';
          _setState(AudioPlaybackState.error);
        }
      }).toJS,
    );

    debugPrint('[Audio] src → ${stream.streamUrl}');
    el.src = stream.streamUrl;
    el.load();
    debugPrint('[Audio] play()');
    el.play().toDart.then((_) {
      debugPrint('[Audio] play() resolved');
    }).catchError((Object e) {
      debugPrint('[Audio] play() rejected: $e');
      if (_version == v && _playbackState != AudioPlaybackState.idle) {
        _errorMessage = 'Lecture bloquée: $e';
        _setState(AudioPlaybackState.error);
      }
      return null;
    });
  }

  void _stopElement() {
    final el = _el;
    if (el != null) {
      el.pause();
      el.src = '';
      _el = null;
    }
  }

  Future<void> pause() async => _el?.pause();

  Future<void> resume() async {
    _el?.play().toDart.catchError((Object _) => null as JSAny?);
  }

  Future<void> stop() async {
    _stopElement();
    _currentStream = null;
    _playbackState = AudioPlaybackState.idle;
    notifyListeners();
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0.0, 1.0);
    if (_el != null) _el!.volume = _volume;
    notifyListeners();
  }

  Future<void> seek(Duration position) async {}
  void toggleShuffle() {}
  void toggleLoop() {}

  Future<void> retry() async {
    if (_currentStream == null) return;
    await playStream(_currentStream!);
  }

  void clearError() {
    _errorMessage = '';
    _setState(AudioPlaybackState.idle);
  }

  void _setState(AudioPlaybackState s) {
    _playbackState = s;
    notifyListeners();
  }

  @override
  void dispose() {
    _stopElement();
    super.dispose();
  }
}
