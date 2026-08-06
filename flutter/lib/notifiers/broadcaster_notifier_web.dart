import 'dart:async';
import 'dart:js_interop';

import 'package:flutter/foundation.dart';
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum BroadcasterState { idle, loading, streaming, error }

class BroadcasterNotifier extends ChangeNotifier {
  static const _flushTimeout = Duration(seconds: 3);
  static const _stopRequestTimeout = Duration(seconds: 7);

  final StreamRepository _repository;

  web.MediaRecorder? _recorder;
  web.MediaStream? _mediaStream;
  Future<void> _chunkUpload = Future<void>.value();
  JSFunction? _pageHideHandler;
  bool _acceptChunks = false;
  bool _disposed = false;

  BroadcasterNotifier(this._repository) {
    _pageHideHandler = ((web.Event _) => _stopForPageHide()).toJS;
    web.window.addEventListener('pagehide', _pageHideHandler);
  }

  BroadcasterState _state = BroadcasterState.idle;
  StreamModel? _currentStream;
  List<StreamModel> _ownedStreams = const [];
  bool _isCatalogLoading = false;
  String _errorMessage = '';

  BroadcasterState get state => _state;
  StreamModel? get currentStream => _currentStream;
  List<StreamModel> get ownedStreams => _ownedStreams;
  bool get isCatalogLoading => _isCatalogLoading;
  String get errorMessage => _errorMessage;
  bool get isStreaming => _state == BroadcasterState.streaming;
  bool get isLoading => _state == BroadcasterState.loading;
  bool get hasError => _state == BroadcasterState.error;

  Future<void> loadOwned() async {
    _isCatalogLoading = true;
    notifyListeners();
    try {
      _ownedStreams = await _repository.getOwned();
    } catch (e) {
      _errorMessage = 'Impossible de charger tes lives: $e';
      _set(BroadcasterState.error);
    } finally {
      _isCatalogLoading = false;
      if (!_disposed) notifyListeners();
    }
  }

  Future<void> startStream(String title) async {
    if (title.isEmpty) {
      _errorMessage = 'Stream title cannot be empty';
      _set(BroadcasterState.error);
      return;
    }
    await _activate(() => _repository.startStream(title));
  }

  Future<void> restartStream(StreamModel stream) async {
    await _activate(() => _repository.restartStream(stream.id));
  }

  Future<void> _activate(Future<StreamModel> Function() activate) async {
    if (isLoading || isStreaming) return;
    _set(BroadcasterState.loading);
    try {
      final stream = await activate();
      if (stream.activeSessionId.isEmpty) {
        throw StateError('Le serveur n\'a pas créé de session de diffusion');
      }
      _currentStream = stream;
      _replaceOwned(stream);
      _errorMessage = '';

      await _startMediaRecorder(stream);
      _set(BroadcasterState.streaming);
    } catch (e) {
      final stream = _currentStream;
      _acceptChunks = false;
      _recorder?.stop();
      _recorder = null;
      _releaseMediaStream();
      if (stream != null) {
        try {
          await _repository.stopStream(stream.id, stream.activeSessionId);
        } catch (_) {}
      }
      if (stream != null) {
        _replaceOwned(stream.copyWith(activeSessionId: '', isLive: false));
      }
      _currentStream = null;
      _errorMessage = e.toString();
      _set(BroadcasterState.error);
    }
  }

  Future<void> _startMediaRecorder(StreamModel stream) async {
    // getUserMedia triggers the browser permission popup. No pre-check needed:
    // hasPermission() returns false for Chrome "prompt" state, causing false
    // negatives. Let the browser handle it.
    try {
      final audioConstraints = web.MediaTrackConstraints(
        echoCancellation: true.toJS,
        noiseSuppression: true.toJS,
        autoGainControl: true.toJS,
        channelCount: 1.toJS,
      );
      final constraints = web.MediaStreamConstraints(audio: audioConstraints);
      _mediaStream = await web.window.navigator.mediaDevices
          .getUserMedia(constraints)
          .toDart;
    } catch (e) {
      throw Exception('getUserMedia: $e');
    }

    // The listener uses MediaSource, so recording and playback must agree on a
    // MIME type that supports incremental WebM/Opus segments.
    const mimeType = 'audio/webm;codecs=opus';
    if (!web.MediaRecorder.isTypeSupported(mimeType) ||
        !web.MediaSource.isTypeSupported(mimeType)) {
      _releaseMediaStream();
      throw Exception(
        'Le streaming WebM/Opus n\'est pas pris en charge par ce navigateur',
      );
    }

    // Retry once after a short delay: on the first cold-load Chrome's audio
    // subsystem occasionally isn't ready yet when MediaRecorder is created.
    web.MediaRecorder? recorder;
    for (var attempt = 0; attempt < 2; attempt++) {
      try {
        recorder = web.MediaRecorder(
          _mediaStream!,
          web.MediaRecorderOptions(mimeType: mimeType),
        );
        break;
      } catch (e) {
        if (attempt == 0) {
          await Future<void>.delayed(const Duration(milliseconds: 300));
        } else {
          throw Exception('MediaRecorder: $e');
        }
      }
    }
    final recorderMimeType = recorder!.mimeType.isEmpty
        ? mimeType
        : recorder.mimeType;
    debugPrint('[Broadcaster] mimeType: $recorderMimeType');

    recorder.onerror = ((web.Event e) {
      debugPrint('[Broadcaster] MediaRecorder error: $e');
      _handleCaptureFailure(
        stream.id,
        stream.activeSessionId,
        'Erreur de capture audio: $e',
      );
    }).toJS;

    final streamId = stream.id;
    final sessionId = stream.activeSessionId;
    _chunkUpload = Future<void>.value();
    _acceptChunks = true;
    recorder.ondataavailable = ((web.BlobEvent event) {
      final blob = event.data;
      debugPrint('[Broadcaster] ondataavailable — blob.size=${blob.size}');
      if (_acceptChunks && blob.size > 0) {
        // Keep both Blob conversion and HTTP upload in the event order. The
        // first WebM blob contains the decoder header; parallel requests could
        // otherwise make a later media fragment reach the server first.
        _chunkUpload = _chunkUpload
            .then((_) async {
              if (!_isCurrentSession(streamId, sessionId)) return;
              final jsBuffer = await blob.arrayBuffer().toDart;
              if (!_isCurrentSession(streamId, sessionId)) return;
              final chunk = jsBuffer.toDart.asUint8List();
              debugPrint('[Broadcaster] pushing chunk: ${chunk.length} bytes');
              await _repository.pushChunk(
                streamId,
                sessionId,
                chunk,
                recorderMimeType,
              );
            })
            .catchError((e) {
              debugPrint('[Broadcaster] chunk upload error: $e');
              _handleCaptureFailure(
                streamId,
                sessionId,
                'Envoi du flux audio interrompu: $e',
              );
            });
      }
    }).toJS;

    recorder.start(500);
    debugPrint('[Broadcaster] recorder.start(500) — state: ${recorder.state}');
    _recorder = recorder;
  }

  bool _isCurrentSession(String streamId, String sessionId) =>
      !_disposed &&
      _acceptChunks &&
      _currentStream?.id == streamId &&
      _currentStream?.activeSessionId == sessionId;

  void _handleCaptureFailure(
    String streamId,
    String sessionId,
    String message,
  ) {
    if (!_isCurrentSession(streamId, sessionId)) return;
    _acceptChunks = false;
    final recorder = _recorder;
    _recorder = null;
    if (recorder != null && recorder.state != 'inactive') {
      recorder.stop();
    }
    _releaseMediaStream();
    final stream = _currentStream;
    _currentStream = null;
    if (stream != null) {
      _replaceOwned(stream.copyWith(activeSessionId: '', isLive: false));
    }
    _errorMessage = message;
    _set(BroadcasterState.error);
    unawaited(_repository.stopStream(streamId, sessionId).catchError((_) {}));
  }

  void _stopForPageHide() {
    final stream = _currentStream;
    if (stream == null) return;
    _acceptChunks = false;
    _recorder?.stop();
    _recorder = null;
    _releaseMediaStream();
    _currentStream = null;
    _replaceOwned(stream.copyWith(activeSessionId: '', isLive: false));
    unawaited(
      _repository
          .stopStream(stream.id, stream.activeSessionId)
          .catchError((_) {}),
    );
  }

  Future<void> stopStream() async {
    final stream = _currentStream;
    if (stream == null) {
      if (_state == BroadcasterState.loading) {
        _set(BroadcasterState.idle);
      }
      return;
    }
    _set(BroadcasterState.loading);
    Object? stopError;
    try {
      try {
        await _stopRecorderAndFlush().timeout(_flushTimeout);
      } on TimeoutException {
        // A suspended chunk request must never keep the studio spinner alive.
        // The request can still finish in the background while /stop closes
        // the server-side publisher.
        debugPrint('[Broadcaster] final audio flush timed out');
      }
      if (_state == BroadcasterState.error) return;
      _acceptChunks = false;
      _releaseMediaStream();
      await _repository
          .stopStream(stream.id, stream.activeSessionId)
          .timeout(_stopRequestTimeout);
    } catch (e) {
      stopError = e;
    } finally {
      _acceptChunks = false;
      final recorder = _recorder;
      _recorder = null;
      if (recorder != null && recorder.state != 'inactive') {
        recorder.stop();
      }
      _releaseMediaStream();
      if (_currentStream?.id == stream.id) {
        _currentStream = null;
      }
      _replaceOwned(stream.copyWith(activeSessionId: '', isLive: false));
      if (!_disposed && _state != BroadcasterState.error) {
        if (stopError == null) {
          _errorMessage = '';
          _set(BroadcasterState.idle);
        } else {
          _errorMessage =
              'Le live a été arrêté localement, mais le serveur '
              'n\'a pas confirmé l\'arrêt: $stopError';
          _set(BroadcasterState.error);
        }
      }
    }
  }

  Future<void> deleteStream(StreamModel stream) async {
    if (_currentStream?.id == stream.id) {
      await stopStream();
    }
    _isCatalogLoading = true;
    notifyListeners();
    try {
      await _repository.deleteStream(stream.id).timeout(_stopRequestTimeout);
      _ownedStreams = _ownedStreams
          .where((item) => item.id != stream.id)
          .toList();
      _errorMessage = '';
      if (_state == BroadcasterState.error) _state = BroadcasterState.idle;
    } catch (e) {
      _errorMessage = 'Impossible de supprimer ce live: $e';
      _state = BroadcasterState.error;
    } finally {
      _isCatalogLoading = false;
      if (!_disposed) notifyListeners();
    }
  }

  void _replaceOwned(StreamModel stream) {
    final streams = [..._ownedStreams];
    final index = streams.indexWhere((item) => item.id == stream.id);
    if (index == -1) {
      streams.insert(0, stream);
    } else {
      streams[index] = stream;
    }
    _ownedStreams = streams;
  }

  Future<void> _stopRecorderAndFlush() async {
    final recorder = _recorder;
    _recorder = null;
    if (recorder == null || recorder.state == 'inactive') {
      await _chunkUpload;
      return;
    }

    final stopped = Completer<void>();
    recorder.addEventListener(
      'stop',
      ((web.Event _) {
        if (!stopped.isCompleted) stopped.complete();
      }).toJS,
    );
    recorder.stop();
    await stopped.future.timeout(const Duration(seconds: 2), onTimeout: () {});
    await _chunkUpload;
  }

  void _releaseMediaStream() {
    final tracks = _mediaStream?.getTracks().toDart;
    if (tracks != null) {
      for (final track in tracks) {
        track.stop();
      }
    }
    _mediaStream = null;
  }

  void clearError() {
    _errorMessage = '';
    _set(BroadcasterState.idle);
  }

  void _set(BroadcasterState s) {
    if (_disposed) return;
    _state = s;
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
    _acceptChunks = false;
    final stream = _currentStream;
    _recorder?.stop();
    _releaseMediaStream();
    if (stream != null) {
      unawaited(
        _repository
            .stopStream(stream.id, stream.activeSessionId)
            .catchError((_) {}),
      );
    }
    super.dispose();
  }
}
