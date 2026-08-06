import 'dart:async';
import 'dart:collection';
import 'dart:js_interop';
import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';
import 'live_playback_policy.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

class AudioNotifier extends ChangeNotifier {
  static const _liveMimeType = 'audio/webm;codecs=opus';
  static const _startupTimeout = Duration(seconds: 12);
  static const _maxLiveLatencySeconds = 2.5;
  static const _retainedBufferedSeconds = 45.0;
  static const _maxBufferedSeconds = 75.0;
  static const _playbackSafetySeconds = 5.0;
  static const _liveEdgeOffsetSeconds = 0.75;
  static const _watchdogInterval = Duration(seconds: 1);
  static const _stalledPlaybackTimeout = Duration(seconds: 5);
  static const _freshBlobWindow = Duration(seconds: 4);
  static const _maxQueuedBlobs = 80;
  static const _reconnectBackoff = <Duration>[
    Duration(milliseconds: 500),
    Duration(seconds: 1),
    Duration(seconds: 2),
    Duration(seconds: 4),
    Duration(seconds: 8),
  ];

  final StreamRepository _repository;
  final Queue<Uint8List> _appendQueue = Queue<Uint8List>();

  web.HTMLAudioElement? _el;
  web.MediaSource? _mediaSource;
  web.SourceBuffer? _sourceBuffer;
  web.WebSocket? _socket;
  Timer? _startupTimer;
  Timer? _playbackWatchdog;
  Timer? _reconnectTimer;
  String? _objectUrl;
  String? _joinedStreamId;
  JSFunction? _pageHideHandler;
  bool _responseEnded = false;
  bool _tearingDown = false;
  bool _disposed = false;
  bool _userPaused = false;
  bool _autoResumeInFlight = false;
  double _lastPlaybackPosition = 0;
  DateTime _lastPlaybackProgressAt = DateTime.fromMillisecondsSinceEpoch(0);
  DateTime _lastBlobAt = DateTime.fromMillisecondsSinceEpoch(0);
  int _receivedBlobCount = 0;
  int _reconnectAttempt = 0;
  int _version = 0;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';

  AudioNotifier({StreamRepository? repository})
    : _repository = repository ?? const StreamRepository() {
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
    _userPaused = false;
    _autoResumeInFlight = false;
    _lastPlaybackPosition = 0;
    _lastPlaybackProgressAt = DateTime.now();
    _lastBlobAt = DateTime.fromMillisecondsSinceEpoch(0);
    _receivedBlobCount = 0;
    _reconnectAttempt = 0;

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
    _startPlaybackWatchdog(el, v);
    _configureMediaSession(stream);
    mediaSource.addEventListener(
      'sourceopen',
      ((web.Event _) {
        if (_version == v) {
          unawaited(_openLiveSocket(stream.streamUrl, mediaSource, v));
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
        if (_isCurrentElement(el, v) && !_userPaused) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'stalled',
      ((web.Event _) {
        if (_isCurrentElement(el, v) && !_userPaused) {
          _setState(AudioPlaybackState.buffering);
        }
      }).toJS,
    );
    el.addEventListener(
      'playing',
      ((web.Event _) {
        if (_isCurrentElement(el, v) && !el.paused) {
          _startupTimer?.cancel();
          _autoResumeInFlight = false;
          _lastPlaybackPosition = el.currentTime;
          _lastPlaybackProgressAt = DateTime.now();
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
          if (_userPaused) {
            _setState(AudioPlaybackState.paused);
          } else {
            debugPrint(
              '[Listener] unexpected pause: ${_playbackDiagnostics(el)}',
            );
            _setState(AudioPlaybackState.buffering);
          }
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

  Future<void> _openLiveSocket(
    String url,
    web.MediaSource mediaSource,
    int v,
  ) async {
    try {
      final sourceBuffer = mediaSource.addSourceBuffer(_liveMimeType);
      // Late listeners resume from a current WebM Cluster after receiving the
      // metadata segment. Sequence mode removes the publisher-uptime gap from
      // the media timestamps and keeps the playback timeline contiguous.
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
            debugPrint('[Listener] SourceBuffer error');
            _fail(v, 'Fragment audio invalide ou reçu dans le désordre');
          }
        }).toJS,
      );

      _connectLiveSocket(url, v);
    } catch (e) {
      _fail(v, 'Impossible de rejoindre le stream: $e');
    }
  }

  void _connectLiveSocket(String url, int v) {
    if (!_canReconnect(v) || _socket != null) return;
    try {
      final socket = web.WebSocket(_webSocketUrl(url));
      socket.binaryType = 'arraybuffer';
      _socket = socket;
      socket.onopen = ((web.Event _) {
        if (_isCurrentSocket(socket, v)) {
          debugPrint('[Listener] WebSocket connected: ${socket.url}');
        }
      }).toJS;
      socket.onmessage = ((web.MessageEvent event) {
        if (!_isCurrentSocket(socket, v)) return;
        final data = event.data;
        if (data == null || !data.isA<JSArrayBuffer>()) {
          _fail(v, 'Le serveur a envoyé un fragment audio non binaire');
          return;
        }
        final bytes = (data as JSArrayBuffer).toDart.asUint8List();
        if (bytes.isEmpty) return;
        _lastBlobAt = DateTime.now();
        _receivedBlobCount++;
        _reconnectAttempt = 0;
        if (_receivedBlobCount <= 3 || _receivedBlobCount % 20 == 0) {
          debugPrint(
            '[Listener] received WebM blob #$_receivedBlobCount: '
            '${bytes.length} bytes queue=${_appendQueue.length}',
          );
        }
        if (_appendQueue.length >= _maxQueuedBlobs) {
          // MediaRecorder produces a dependent byte stream. Dropping one
          // arbitrary message would corrupt every later append, so reconnect
          // with a fresh MediaSource instead.
          debugPrint('[Listener] append queue overflow; rebuilding playback');
          _socket = null;
          _closeSocket(socket, 'append queue overflow');
          unawaited(_handleSocketClosed(v));
          return;
        }
        _appendQueue.add(bytes);
        _appendNext(v);
      }).toJS;
      socket.onerror = ((web.Event e) {
        if (_isCurrentSocket(socket, v)) {
          // Browsers report the useful close code through onclose. A temporary
          // socket error must not destroy an otherwise healthy multi-hour
          // listening session.
          debugPrint('[Listener] WebSocket error; waiting for close: $e');
        }
      }).toJS;
      socket.onclose = ((web.CloseEvent event) {
        if (!_isCurrentSocket(socket, v)) return;
        debugPrint(
          '[Listener] WebSocket closed: code=${event.code} '
          'reason=${event.reason} clean=${event.wasClean}',
        );
        _socket = null;
        unawaited(_handleSocketClosed(v));
      }).toJS;
    } catch (e) {
      debugPrint('[Listener] WebSocket connection failed: $e');
      final stream = _currentStream;
      if (stream != null) _schedulePipelineRestart(stream, v);
    }
  }

  Future<void> _handleSocketClosed(int v) async {
    if (!_canReconnect(v)) return;
    if (_userPaused) {
      ++_version;
      _stopElement();
      _setState(AudioPlaybackState.paused);
      return;
    }
    _setState(AudioPlaybackState.buffering);

    final blobsAtClose = _receivedBlobCount;
    final streamId = _currentStream?.id;
    if (streamId == null) return;
    try {
      var activeStreams = await _repository.getActive().timeout(
        const Duration(seconds: 5),
      );
      if (!_canReconnect(v) || _receivedBlobCount != blobsAtClose) return;
      var activeStream = _streamById(activeStreams, streamId);
      if (activeStream == null) {
        // Confirm an ended live twice so a short database/API transition does
        // not terminate a listener which could otherwise reconnect.
        await Future<void>.delayed(const Duration(seconds: 1));
        if (!_canReconnect(v) || _receivedBlobCount != blobsAtClose) return;
        activeStreams = await _repository.getActive().timeout(
          const Duration(seconds: 5),
        );
        if (!_canReconnect(v) || _receivedBlobCount != blobsAtClose) return;
        activeStream = _streamById(activeStreams, streamId);
      }
      if (activeStream != null) {
        final previousSession = _currentStream?.activeSessionId ?? '';
        final activeSession = activeStream.activeSessionId;
        if (previousSession.isNotEmpty &&
            activeSession.isNotEmpty &&
            previousSession != activeSession) {
          debugPrint(
            '[Listener] stream restarted with a new session; rebuilding MSE',
          );
        } else {
          debugPrint(
            '[Listener] WebSocket interrupted; rebuilding MSE at a Cluster',
          );
        }
        _schedulePipelineRestart(activeStream, v);
        return;
      }

      debugPrint('[Listener] stream is no longer live; finishing playback');
      _responseEnded = true;
      _reconnectTimer?.cancel();
      _reconnectTimer = null;
      final socket = _socket;
      _socket = null;
      _closeSocket(socket, 'stream ended');
      _finishMediaSourceIfReady(v);
    } catch (e) {
      // If the REST health check is unavailable, recover the media pipeline.
      // A temporary control-plane outage must not terminate playback.
      debugPrint('[Listener] live status check failed; rebuilding MSE: $e');
      final stream = _currentStream;
      if (stream != null) _schedulePipelineRestart(stream, v);
    }
  }

  StreamModel? _streamById(List<StreamModel> streams, String streamId) {
    for (final stream in streams) {
      if (stream.id == streamId) return stream;
    }
    return null;
  }

  void _schedulePipelineRestart(StreamModel stream, int v) {
    if (!_canReconnect(v) || _reconnectTimer?.isActive == true) return;
    final index = math.min(_reconnectAttempt, _reconnectBackoff.length - 1);
    final delay = _reconnectBackoff[index];
    _reconnectAttempt++;
    debugPrint(
      '[Listener] rebuilding playback in ${delay.inMilliseconds}ms '
      '(attempt $_reconnectAttempt)',
    );
    _reconnectTimer = Timer(delay, () {
      _reconnectTimer = null;
      if (!_canReconnect(v)) return;
      unawaited(playStream(stream));
    });
  }

  bool _canReconnect(int v) =>
      !_disposed &&
      !_tearingDown &&
      !_responseEnded &&
      _version == v &&
      _el != null &&
      _sourceBuffer != null;

  String _webSocketUrl(String url) {
    final uri = Uri.parse(url);
    return uri
        .replace(
          scheme: uri.scheme == 'https' ? 'wss' : 'ws',
          path: '${uri.path}/ws',
        )
        .toString();
  }

  bool _isCurrentSocket(web.WebSocket socket, int v) =>
      !_disposed && !_tearingDown && _version == v && _socket == socket;

  void _closeSocket(web.WebSocket? socket, String reason) {
    if (socket == null || socket.readyState >= web.WebSocket.CLOSING) return;
    try {
      socket.close(1000, reason);
    } catch (e) {
      // Chromium may reject close() while the WebSocket handshake is still
      // pending. Clearing our reference is enough for stale callbacks to be
      // ignored; the browser will release the failed connection itself.
      debugPrint('[Listener] WebSocket close ignored: $e');
    }
  }

  void _appendNext(int v) {
    if (_version != v) return;
    final sourceBuffer = _sourceBuffer;
    if (sourceBuffer == null || sourceBuffer.updating) return;
    // Incoming audio always has priority over maintenance removals. Trimming
    // before every append eventually starves the queue during long sessions.
    if (_appendQueue.isNotEmpty) {
      final bytes = _appendQueue.removeFirst();
      try {
        sourceBuffer.appendBuffer(bytes.toJS);
      } catch (e) {
        _fail(v, 'Impossible de mettre en mémoire le stream: $e');
      }
      return;
    }
    if (_trimLiveBuffer(sourceBuffer)) return;
    _finishMediaSourceIfReady(v);
  }

  bool _trimLiveBuffer(web.SourceBuffer sourceBuffer) {
    try {
      final buffered = sourceBuffer.buffered;
      if (buffered.length == 0) return false;
      if (_appendQueue.length > 2) return false;
      final bufferStart = buffered.start(0);
      final liveEnd = buffered.end(buffered.length - 1);
      if (liveEnd - bufferStart < _maxBufferedSeconds) return false;
      final desiredRemoveEnd = liveEnd - _retainedBufferedSeconds;
      final el = _el;
      final playbackSafeEnd = el == null || _userPaused
          ? desiredRemoveEnd
          : el.currentTime - _playbackSafetySeconds;
      final removeEnd = math.min(desiredRemoveEnd, playbackSafeEnd);
      if (removeEnd > bufferStart + 1) {
        debugPrint(
          '[Listener] trimming buffer: ${bufferStart.toStringAsFixed(1)}-'
          '${removeEnd.toStringAsFixed(1)} live=${liveEnd.toStringAsFixed(1)} '
          'current=${el?.currentTime.toStringAsFixed(1) ?? 'none'}',
        );
        sourceBuffer.remove(bufferStart, removeEnd);
        return true;
      }
    } catch (_) {
      // Some browsers expose a transient empty TimeRanges while updating.
    }
    return false;
  }

  void _startPlaybackWatchdog(web.HTMLAudioElement el, int v) {
    _playbackWatchdog?.cancel();
    _playbackWatchdog = Timer.periodic(_watchdogInterval, (_) {
      if (!_isCurrentElement(el, v) || _userPaused) return;
      final now = DateTime.now();
      final position = el.currentTime;
      if ((position - _lastPlaybackPosition).abs() > 0.05) {
        _lastPlaybackPosition = position;
        _lastPlaybackProgressAt = now;
        return;
      }
      final blobsAreFresh = now.difference(_lastBlobAt) <= _freshBlobWindow;
      final playbackIsStalled =
          now.difference(_lastPlaybackProgressAt) >= _stalledPlaybackTimeout;
      if (!blobsAreFresh || !playbackIsStalled) return;

      debugPrint(
        '[Listener] playback watchdog recovery: ${_playbackDiagnostics(el)}',
      );
      _lastPlaybackProgressAt = now;
      _seekToLiveEdge(allowWhilePaused: true);
      unawaited(_recoverUnexpectedPause(el, v));
    });
  }

  Future<void> _recoverUnexpectedPause(web.HTMLAudioElement el, int v) async {
    if (_autoResumeInFlight || _userPaused || !_isCurrentElement(el, v)) return;
    _autoResumeInFlight = true;
    try {
      _seekToLiveEdge(allowWhilePaused: true);
      await el.play().toDart;
      if (_isCurrentElement(el, v) && !el.paused) {
        _lastPlaybackPosition = el.currentTime;
        _lastPlaybackProgressAt = DateTime.now();
        _setState(AudioPlaybackState.playing);
      }
    } catch (e) {
      if (_isCurrentElement(el, v)) {
        debugPrint('[Listener] automatic resume failed: $e');
      }
    } finally {
      if (_version == v) _autoResumeInFlight = false;
    }
  }

  String _playbackDiagnostics(web.HTMLAudioElement el) {
    var buffered = 'none';
    var bufferedAhead = 0.0;
    try {
      final ranges = el.buffered;
      if (ranges.length > 0) {
        final liveEnd = ranges.end(ranges.length - 1);
        bufferedAhead = liveEnd - el.currentTime;
        buffered =
            '${ranges.start(0).toStringAsFixed(2)}-'
            '${liveEnd.toStringAsFixed(2)}';
      }
    } catch (_) {}
    return 'paused=${el.paused} current=${el.currentTime.toStringAsFixed(2)} '
        'buffered=$buffered ahead=${bufferedAhead.toStringAsFixed(2)} '
        'readyState=${el.readyState} queue=${_appendQueue.length} '
        'sourceUpdating=${_sourceBuffer?.updating} socket=${_socket?.readyState}';
  }

  void _seekToLiveEdge({bool allowWhilePaused = false}) {
    final el = _el;
    if (el == null || (el.paused && !allowWhilePaused)) return;
    try {
      final buffered = el.buffered;
      if (buffered.length == 0) return;
      final lastRange = buffered.length - 1;
      final liveStart = buffered.start(lastRange);
      final liveEnd = buffered.end(lastRange);
      final currentTime = el.currentTime;
      final target = liveSeekTarget(
        currentTime: currentTime,
        rangeStart: liveStart,
        rangeEnd: liveEnd,
        maxLatency: _maxLiveLatencySeconds,
        edgeOffset: _liveEdgeOffsetSeconds,
      );
      if (target == null) return;

      debugPrint(
        '[Listener] forward live seek: ${currentTime.toStringAsFixed(2)} -> '
        '${target.toStringAsFixed(2)} buffered='
        '${liveStart.toStringAsFixed(2)}-${liveEnd.toStringAsFixed(2)}',
      );
      el.currentTime = target;
      _lastPlaybackPosition = target;
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
    debugPrint('[Listener] playback failure: $message');
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
    _playbackWatchdog?.cancel();
    _playbackWatchdog = null;
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempt = 0;
    final socket = _socket;
    _socket = null;
    _closeSocket(socket, 'client stop');
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
    _userPaused = true;
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
    _userPaused = false;
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
    _userPaused = false;
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
    super.dispose();
  }
}
