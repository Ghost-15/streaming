import 'dart:async';
import 'dart:js_interop';

import 'package:flutter/foundation.dart';
import 'package:web/web.dart' as web;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum BroadcasterState { idle, loading, streaming, error }

class BroadcasterNotifier extends ChangeNotifier {
  final StreamRepository _repository;

  web.MediaRecorder? _recorder;
  web.MediaStream? _mediaStream;
  Future<void> _chunkUpload = Future<void>.value();

  BroadcasterNotifier(this._repository);

  BroadcasterState _state = BroadcasterState.idle;
  StreamModel? _currentStream;
  String _errorMessage = '';
  int _listenerCount = 0;

  BroadcasterState get state => _state;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;
  int get listenerCount => _listenerCount;
  bool get isStreaming => _state == BroadcasterState.streaming;
  bool get isLoading => _state == BroadcasterState.loading;
  bool get hasError => _state == BroadcasterState.error;

  Future<void> startStream(String title) async {
    if (title.isEmpty) {
      _errorMessage = 'Stream title cannot be empty';
      _set(BroadcasterState.error);
      return;
    }
    _set(BroadcasterState.loading);
    try {
      final stream = await _repository.startStream(title);
      _currentStream = stream;
      _listenerCount = 0;
      _errorMessage = '';

      await _startMediaRecorder(stream);
      _set(BroadcasterState.streaming);
    } catch (e) {
      final stream = _currentStream;
      _recorder?.stop();
      _recorder = null;
      _releaseMediaStream();
      if (stream != null) {
        try {
          await _repository.stopStream(stream.id);
        } catch (_) {}
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
      final constraints = web.MediaStreamConstraints(audio: true.toJS);
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
    final recorderMimeType = recorder!.mimeType;
    debugPrint('[Broadcaster] mimeType: $recorderMimeType');

    recorder.onerror = ((web.Event e) {
      debugPrint('[Broadcaster] MediaRecorder error: $e');
    }).toJS;

    final streamId = stream.id;
    _chunkUpload = Future<void>.value();
    recorder.ondataavailable = ((web.BlobEvent event) {
      final blob = event.data;
      debugPrint('[Broadcaster] ondataavailable — blob.size=${blob.size}');
      if (blob.size > 0) {
        // Keep both Blob conversion and HTTP upload in the event order. The
        // first WebM blob contains the decoder header; parallel requests could
        // otherwise make a later media fragment reach the server first.
        _chunkUpload = _chunkUpload
            .then((_) async {
              final jsBuffer = await blob.arrayBuffer().toDart;
              final chunk = jsBuffer.toDart.asUint8List();
              debugPrint('[Broadcaster] pushing chunk: ${chunk.length} bytes');
              await _repository.pushChunk(streamId, chunk, recorderMimeType);
            })
            .catchError((e) {
              debugPrint('[Broadcaster] chunk upload error: $e');
            });
      }
    }).toJS;

    recorder.start(250);
    debugPrint('[Broadcaster] recorder.start(250) — state: ${recorder.state}');
    _recorder = recorder;
  }

  Future<void> stopStream() async {
    if (_currentStream == null) return;
    _set(BroadcasterState.loading);
    try {
      await _stopRecorderAndFlush();
      _releaseMediaStream();
      await _repository.stopStream(_currentStream!.id);
      _currentStream = null;
      _listenerCount = 0;
      _errorMessage = '';
      _set(BroadcasterState.idle);
    } catch (e) {
      _errorMessage = e.toString();
      _set(BroadcasterState.error);
    }
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

  void updateListenerCount(int count) {
    _listenerCount = count;
    notifyListeners();
  }

  void clearError() {
    _errorMessage = '';
    _set(BroadcasterState.idle);
  }

  void _set(BroadcasterState s) {
    _state = s;
    notifyListeners();
  }

  @override
  void dispose() {
    _recorder?.stop();
    _releaseMediaStream();
    super.dispose();
  }
}
