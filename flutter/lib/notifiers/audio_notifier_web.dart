import 'dart:async';
import 'dart:collection';
import 'dart:js_interop';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

class AudioNotifier extends ChangeNotifier {
  static const _liveMimeType = 'audio/webm;codecs=opus';
  static const _startupTimeout = Duration(seconds: 12);
  static const _maxLiveLatencySeconds = 2.0;
  static const _retainedBufferedSeconds = 15.0;
  static const _liveEdgeOffsetSeconds = 0.25;

  final http.Client _streamClient;
  final StreamRepository _repository;
  final Queue<Uint8List> _appendQueue = Queue<Uint8List>();

  web.HTMLAudioElement? _el;
  web.MediaSource? _mediaSource;
  web.SourceBuffer? _sourceBuffer;
  StreamSubscription<List<int>>? _audioSubscription;
  Timer? _startupTimer;
  String? _objectUrl;
  String? _joinedStreamId;
  JSFunction? _pageHideHandler;
  bool _responseEnded = false;
  bool _tearingDown = false;
  bool _disposed = false;
  int _version = 0;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';

  AudioNotifier({http.Client? streamClient, StreamRepository? repository})
    : _streamClient = streamClient ?? http.Client(),
      _repository = repository ?? const StreamRepository() {
    _pageHideHandler = ((web.Event _) => _leaveJoinedStream()).toJS;
    web.window.addEventListener('pagehide', _pageHideHandler);
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

  Future<void> playStream(StreamModel stream) async {
    final v = ++_version;
    _leaveJoinedStream();
    _stopElement();

    _playbackState = AudioPlaybackState.loading;
    _currentStream = stream;
    _errorMessage = '';
    notifyListeners();

    unawaited(_joinStream(stream.id, v));

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
    el.preload = 'auto';
    el.setAttribute('playsinline', '');

    _bindElementEvents(el, v);
    _configureMediaSession(stream);
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

  Future<void> _joinStream(String streamId, int v) async {
    try {
      await _repository.joinStream(streamId);
      if (_disposed ||
          _version != v ||
          _currentStream?.id != streamId ||
          _playbackState == AudioPlaybackState.error ||
          _playbackState == AudioPlaybackState.idle) {
        unawaited(_repository.leaveStream(streamId).catchError((_) {}));
        return;
      }
      _joinedStreamId = streamId;
    } catch (_) {
      // The public audio endpoint remains usable for anonymous listeners. The
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

  void _bindElementEvents(web.HTMLAudioElement el, int v) {
    el.addEventListener(
      'waiting',
      ((web.Event _) {
        if (_isCurrentElement(el, v) &&
            !el.paused &&
            _playbackState == AudioPlaybackState.playing) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'stalled',
      ((web.Event _) {
        if (_isCurrentElement(el, v) &&
            !el.paused &&
            _playbackState == AudioPlaybackState.playing) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'playing',
      ((web.Event _) {
        if (_isCurrentElement(el, v) && !el.paused) {
          _startupTimer?.cancel();
          _setState(AudioPlaybackState.playing);
        }
      }).toJS,
    );
    el.addEventListener(
      'pause',
      ((web.Event _) {
        if (_isCurrentElement(el, v) &&
            _playbackState != AudioPlaybackState.idle &&
            _playbackState != AudioPlaybackState.error) {
          _setState(AudioPlaybackState.paused);
        }
      }).toJS,
    );
    el.addEventListener(
      'ended',
      ((web.Event _) {
        if (_isCurrentElement(el, v)) {
          _stopElement();
          _leaveJoinedStream();
          _currentStream = null;
          _setState(AudioPlaybackState.idle);
          _clearMediaSession();
        }
      }).toJS,
    );
    el.addEventListener(
      'error',
      ((web.Event _) {
        if (_isCurrentElement(el, v)) {
          _fail(v, 'Erreur de décodage du stream audio');
        }
      }).toJS,
    );
    el.addEventListener(
      'abort',
      ((web.Event _) {
        if (_isCurrentElement(el, v)) {
          _fail(v, 'La lecture du stream a été interrompue');
        }
      }).toJS,
    );
  }

  bool _isCurrentElement(web.HTMLAudioElement el, int v) =>
      !_disposed && !_tearingDown && _version == v && _el == el;

  Future<void> _openLiveResponse(
    String url,
    web.MediaSource mediaSource,
    int v,
  ) async {
    try {
      final request = http.Request('GET', Uri.parse(url));
      final response = await _streamClient.send(request);
      if (_version != v) return;
      if (response.statusCode != 200) {
        throw Exception('HTTP ${response.statusCode}');
      }

      final responseType = response.headers['content-type'] ?? _liveMimeType;
      final mimeType = web.MediaSource.isTypeSupported(responseType)
          ? responseType
          : _liveMimeType;
      final sourceBuffer = mediaSource.addSourceBuffer(mimeType);
      sourceBuffer.mode = 'sequence';
      _sourceBuffer = sourceBuffer;

      sourceBuffer.addEventListener(
        'updateend',
        ((web.Event _) {
          if (_version == v) {
            _seekToLiveEdge();
            _appendNext(v);
          }
        }).toJS,
      );
      sourceBuffer.addEventListener(
        'error',
        ((web.Event _) {
          if (_version == v) {
            _fail(v, 'Fragment audio invalide ou reçu dans le désordre');
          }
        }).toJS,
      );

      _audioSubscription = response.stream.listen(
        (bytes) {
          if (_version != v || bytes.isEmpty) return;
          _appendQueue.add(Uint8List.fromList(bytes));
          _appendNext(v);
        },
        onError: (Object e) => _fail(v, 'Connexion au stream interrompue: $e'),
        onDone: () {
          if (_version != v) return;
          _responseEnded = true;
          _finishMediaSourceIfReady(v);
        },
        cancelOnError: true,
      );
    } catch (e) {
      _fail(v, 'Impossible de rejoindre le stream: $e');
    }
  }

  void _appendNext(int v) {
    if (_version != v) return;
    final sourceBuffer = _sourceBuffer;
    if (sourceBuffer == null || sourceBuffer.updating) return;
    if (_trimLiveBuffer(sourceBuffer)) return;
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

  bool _trimLiveBuffer(web.SourceBuffer sourceBuffer) {
    try {
      final buffered = sourceBuffer.buffered;
      if (buffered.length == 0) return false;
      final liveEnd = buffered.end(buffered.length - 1);
      final removeEnd = liveEnd - _retainedBufferedSeconds;
      if (removeEnd > buffered.start(0)) {
        sourceBuffer.remove(0, removeEnd);
        return true;
      }
    } catch (_) {
      // Some browsers expose a transient empty TimeRanges while updating.
    }
    return false;
  }

  void _seekToLiveEdge({bool allowWhilePaused = false}) {
    final el = _el;
    if (el == null || (el.paused && !allowWhilePaused)) return;
    try {
      final buffered = el.buffered;
      if (buffered.length == 0) return;
      final liveEnd = buffered.end(buffered.length - 1);
      if (liveEnd - el.currentTime > _maxLiveLatencySeconds) {
        el.currentTime = liveEnd - _liveEdgeOffsetSeconds;
      }
    } catch (_) {}
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
    // Invalidate callbacks bound to the failed element before tearing it down.
    ++_version;
    _errorMessage = message;
    _playbackState = AudioPlaybackState.error;
    _stopElement();
    _leaveJoinedStream();
    _syncMediaSessionState();
    notifyListeners();
  }

  void _stopElement() {
    _tearingDown = true;
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
    _tearingDown = false;
  }

  Future<void> pause() async {
    final el = _el;
    if (el == null) return;
    final v = _version;
    el.pause();
    if (_isCurrentElement(el, v) &&
        _playbackState != AudioPlaybackState.idle &&
        _playbackState != AudioPlaybackState.error) {
      _setState(AudioPlaybackState.paused);
    }
  }

  Future<void> resume() async {
    final el = _el;
    if (el == null) {
      final stream = _currentStream;
      if (stream != null) await playStream(stream);
      return;
    }
    final v = _version;
    _seekToLiveEdge(allowWhilePaused: true);
    try {
      await el.play().toDart;
      if (_isCurrentElement(el, v)) _seekToLiveEdge();
    } catch (e) {
      if (_isCurrentElement(el, v)) {
        _fail(v, 'Lecture bloquée par le navigateur: $e');
      }
    }
  }

  Future<void> stop() async {
    ++_version;
    _stopElement();
    _leaveJoinedStream();
    _currentStream = null;
    _playbackState = AudioPlaybackState.idle;
    _clearMediaSession();
    notifyListeners();
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0.0, 1.0);
    if (_el != null) _el!.volume = _volume;
    notifyListeners();
  }

  void _configureMediaSession(StreamModel stream) {
    try {
      final session = web.window.navigator.mediaSession;
      session.metadata = web.MediaMetadata(
        web.MediaMetadataInit(
          title: stream.title,
          artist: stream.broadcasterName.isEmpty
              ? 'StreamPulse'
              : stream.broadcasterName,
          album: 'Direct StreamPulse',
        ),
      );
      _setMediaSessionAction(session, 'play', (() => unawaited(resume())).toJS);
      _setMediaSessionAction(session, 'pause', (() => unawaited(pause())).toJS);
      _setMediaSessionAction(session, 'stop', (() => unawaited(stop())).toJS);
      _syncMediaSessionState();
    } catch (_) {
      // Media Session is optional (notably absent from Firefox). Playback via
      // HTMLAudioElement continues to work when system controls are unavailable.
    }
  }

  void _setMediaSessionAction(
    web.MediaSession session,
    String action,
    JSFunction? handler,
  ) {
    try {
      session.setActionHandler(action, handler);
    } catch (_) {
      // Browsers may expose Media Session but reject individual actions.
    }
  }

  void _syncMediaSessionState() {
    try {
      web.window.navigator.mediaSession.playbackState =
          switch (_playbackState) {
            AudioPlaybackState.playing ||
            AudioPlaybackState.loading ||
            AudioPlaybackState.buffering => 'playing',
            AudioPlaybackState.paused => 'paused',
            _ => 'none',
          };
    } catch (_) {}
  }

  void _clearMediaSession() {
    try {
      final session = web.window.navigator.mediaSession;
      _setMediaSessionAction(session, 'play', null);
      _setMediaSessionAction(session, 'pause', null);
      _setMediaSessionAction(session, 'stop', null);
      session.metadata = null;
      session.playbackState = 'none';
    } catch (_) {}
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
    if (_disposed) return;
    _playbackState = state;
    _syncMediaSessionState();
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    final pageHideHandler = _pageHideHandler;
    if (pageHideHandler != null) {
      web.window.removeEventListener('pagehide', pageHideHandler);
      _pageHideHandler = null;
    }
    ++_version;
    _stopElement();
    _leaveJoinedStream();
    _clearMediaSession();
    _streamClient.close();
    super.dispose();
  }
}
