import 'dart:async';
import 'dart:collection';
import 'dart:js_interop';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../services/storage_service.dart';

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
  static const _liveMimeType = 'audio/webm;codecs=opus';
  static const _startupTimeout = Duration(seconds: 12);

  final http.Client _streamClient;
  final Queue<Uint8List> _appendQueue = Queue<Uint8List>();

  web.HTMLAudioElement? _el;
  web.MediaSource? _mediaSource;
  web.SourceBuffer? _sourceBuffer;
  StreamSubscription<List<int>>? _audioSubscription;
  Timer? _startupTimer;
  String? _objectUrl;
  bool _responseEnded = false;
  int _version = 0;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';
  bool _isShuffled = false;
  bool _isLooping = false;

  AudioNotifier({http.Client? streamClient})
    : _streamClient = streamClient ?? http.Client();

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  Duration get position => Duration.zero;
  Duration get duration => Duration.zero;
  double get progress => 0.0;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;
  bool get isShuffled => _isShuffled;
  bool get isLooping => _isLooping;

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
    _errorMessage = '';
    notifyListeners();

    const StreamRepository().joinStream(stream.id).catchError((_) {});

    if (!web.MediaSource.isTypeSupported(_liveMimeType)) {
      _fail(
        v,
        'Le streaming audio n\'est pas pris en charge par ce navigateur',
      );
      return;
    }

    final el = web.HTMLAudioElement();
    final mediaSource = web.MediaSource();
    final objectUrl = web.URL.createObjectURL(mediaSource);
    _el = el;
    _mediaSource = mediaSource;
    _objectUrl = objectUrl;
    el.volume = _volume;
    el.crossOrigin = 'anonymous';

    _bindElementEvents(el, v);
    mediaSource.addEventListener(
      'sourceopen',
      ((web.Event _) {
        if (_version == v) {
          unawaited(_openLiveResponse(stream.streamUrl, mediaSource, v));
        }
      }).toJS,
    );

    el.src = objectUrl;
    el.load();

    // Keep play() in the original click call stack so browser autoplay rules
    // recognise it as a user action. The promise resolves after data is added.
    el.play().toDart.catchError((Object e) {
      if (_version == v && _playbackState != AudioPlaybackState.idle) {
        _fail(v, 'Lecture bloquée par le navigateur: $e');
      }
      return null;
    });

    _startupTimer = Timer(_startupTimeout, () {
      if (_version == v && !isPlaying) {
        _fail(v, 'Le stream ne fournit aucune donnée audio lisible');
      }
    });
  }

  void _bindElementEvents(web.HTMLAudioElement el, int v) {
    el.addEventListener(
      'waiting',
      ((web.Event _) {
        if (_version == v && !isLoading) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'stalled',
      ((web.Event _) {
        if (_version == v && !isLoading) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'playing',
      ((web.Event _) {
        if (_version == v) {
          _startupTimer?.cancel();
          _setState(AudioPlaybackState.playing);
        }
      }).toJS,
    );
    el.addEventListener(
      'pause',
      ((web.Event _) {
        if (_version == v &&
            _playbackState != AudioPlaybackState.idle &&
            _playbackState != AudioPlaybackState.error) {
          _setState(AudioPlaybackState.paused);
        }
      }).toJS,
    );
    el.addEventListener(
      'ended',
      ((web.Event _) {
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
        if (_version == v) {
          _fail(v, 'Erreur de décodage du stream audio');
        }
      }).toJS,
    );
  }

  Future<void> _openLiveResponse(
    String url,
    web.MediaSource mediaSource,
    int v,
  ) async {
    try {
      _playbackState = AudioPlaybackState.loading;
      _currentStream = stream;
      notifyListeners();

      final token = await StorageService.get(StorageKey.token);
      await _audioPlayer.setUrl(
        stream.streamUrl,
        headers: {if (token != null) 'Authorization': 'Bearer $token'},
      );
    } catch (e) {
      _fail(v, 'Impossible de rejoindre le stream: $e');
    }
  }

  void _appendNext(int v) {
    if (_version != v) return;
    final sourceBuffer = _sourceBuffer;
    if (sourceBuffer == null || sourceBuffer.updating) return;
    if (_appendQueue.isEmpty) {
      _finishMediaSourceIfReady(v);
      return;
    }

    final bytes = _appendQueue.removeFirst();
    try {
      sourceBuffer.appendBuffer(bytes.toJS);
    } catch (e) {
      _fail(v, 'Impossible de mettre en mémoire le stream: $e');
    }
  }

  void _finishMediaSourceIfReady(int v) {
    if (_version != v || !_responseEnded || _appendQueue.isNotEmpty) return;
    final sourceBuffer = _sourceBuffer;
    final mediaSource = _mediaSource;
    if (sourceBuffer == null || sourceBuffer.updating || mediaSource == null) {
      return;
    }
    if (mediaSource.readyState == 'open') {
      mediaSource.endOfStream();
    }
  }

  void _fail(int v, String message) {
    if (_version != v) return;
    _errorMessage = message;
    _playbackState = AudioPlaybackState.error;
    _startupTimer?.cancel();
    unawaited(_audioSubscription?.cancel());
    notifyListeners();
  }

  void _stopElement() {
    _startupTimer?.cancel();
    _startupTimer = null;
    unawaited(_audioSubscription?.cancel());
    _audioSubscription = null;
    _appendQueue.clear();
    _responseEnded = false;
    _sourceBuffer = null;
    _mediaSource = null;

    final el = _el;
    if (el != null) {
      el.pause();
      el.removeAttribute('src');
      el.load();
      _el = null;
    }
    final objectUrl = _objectUrl;
    if (objectUrl != null) {
      web.URL.revokeObjectURL(objectUrl);
      _objectUrl = null;
    }
  }

  Future<void> pause() async => _el?.pause();

  Future<void> resume() async {
    _el?.play().toDart.catchError((Object _) => null);
  }

  Future<void> stop() async {
    ++_version;
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

  void toggleShuffle() {
    _isShuffled = !_isShuffled;
    notifyListeners();
  }

  void toggleLoop() {
    _isLooping = !_isLooping;
    notifyListeners();
  }

  Future<void> retry() async {
    if (_currentStream == null) return;
    await playStream(_currentStream!);
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
    ++_version;
    _stopElement();
    _streamClient.close();
    super.dispose();
  }
}
