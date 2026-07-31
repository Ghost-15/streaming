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

    // No mimeType override: Chrome defaults to audio/webm;codecs=opus for
    // audio-only streams. Forcing the MIME string can throw OperationError
    // if the exact codec label doesn't match the browser's internal list.
    // Retry once after a short delay: on the first cold-load Chrome's audio
    // subsystem occasionally isn't ready yet when MediaRecorder is created.
    web.MediaRecorder? recorder;
    for (var attempt = 0; attempt < 2; attempt++) {
      try {
        recorder = web.MediaRecorder(_mediaStream!);
        break;
      } catch (e) {
        if (attempt == 0) {
          await Future<void>.delayed(const Duration(milliseconds: 300));
        } else {
          throw Exception('MediaRecorder: $e');
        }
      }
    }
    final mimeType = recorder!.mimeType;
    debugPrint('[Broadcaster] mimeType: $mimeType');

    recorder.onerror = ((web.Event e) {
      debugPrint('[Broadcaster] MediaRecorder error: $e');
    }).toJS;

    final streamId = stream.id;
    recorder.ondataavailable = ((web.BlobEvent event) {
      final blob = event.data;
      debugPrint('[Broadcaster] ondataavailable — blob.size=${blob.size}');
      if (blob.size > 0) {
        blob.arrayBuffer().toDart.then((jsBuffer) {
          final chunk = jsBuffer.toDart.asUint8List();
          debugPrint('[Broadcaster] pushing chunk: ${chunk.length} bytes');
          _repository.pushChunk(streamId, chunk, mimeType).catchError((e) {
            debugPrint('[Broadcaster] pushChunk error: $e');
          });
        }).catchError((e) {
          debugPrint('[Broadcaster] arrayBuffer error: $e');
        });
      }
    }).toJS;

    recorder.start(100); // 100 ms timeslice → one WebM cluster per chunk
    debugPrint('[Broadcaster] recorder.start(100) — state: ${recorder.state}');
    _recorder = recorder;
  }

  Future<void> stopStream() async {
    if (_currentStream == null) return;
    _set(BroadcasterState.loading);
    try {
      _recorder?.stop();
      _recorder = null;
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
