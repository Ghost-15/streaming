import 'dart:async';
import 'dart:typed_data';

import 'package:flutter/foundation.dart';
import 'package:record/record.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum BroadcasterState { idle, loading, streaming, error }

/// Mobile (Android/iOS) microphone broadcaster.
///
/// Captures the microphone with `record` and pushes the encoded audio to
/// POST /streams/:id/push, mirroring the Web MediaRecorder flow (one flush every
/// ~500 ms, ordered, with the active broadcast session id). Public surface is
/// identical to the stub and Web notifiers so the UI stays platform-agnostic.
///
/// Note on codecs: mobile capture uses AAC (self-framed ADTS), not the browser's
/// WebM/Opus. It therefore interoperates with the mobile listeners (just_audio on
/// Android, media_kit on iOS) but not with a browser <audio> element.
class BroadcasterNotifier extends ChangeNotifier {
  BroadcasterNotifier(this._repository);

  static const _mimeType = 'audio/aac';
  static const _flushInterval = Duration(milliseconds: 500);

  final StreamRepository _repository;
  // Created lazily on first capture so merely constructing the notifier (unit
  // tests, providers) never touches the microphone platform channel.
  AudioRecorder? _recorder;
  StreamSubscription<Uint8List>? _captureSub;
  Timer? _flushTimer;
  final BytesBuilder _buffer = BytesBuilder(copy: false);
  Future<void> _pushChain = Future<void>.value();

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
    } finally {
      _isCatalogLoading = false;
      notifyListeners();
    }
  }

  Future<void> startStream(String title) async {
    if (title.trim().isEmpty) {
      _errorMessage = 'Stream title cannot be empty';
      _set(BroadcasterState.error);
      return;
    }
    await _activate(() => _repository.startStream(title));
  }

  Future<void> restartStream(StreamModel stream) async {
    await _activate(() => _repository.restartStream(stream.id));
  }

  Future<void> _activate(Future<StreamModel> Function() create) async {
    _set(BroadcasterState.loading);
    StreamModel? stream;
    try {
      stream = await create();
      _currentStream = stream;
      if (stream.activeSessionId.isEmpty) {
        throw Exception('Session de diffusion manquante');
      }
      await _startCapture(stream);
      _errorMessage = '';
      _set(BroadcasterState.streaming);
    } catch (e) {
      await _stopCapture();
      final started = stream ?? _currentStream;
      if (started != null) {
        try {
          await _repository.stopStream(started.id, started.activeSessionId);
        } catch (_) {}
      }
      _currentStream = null;
      _errorMessage = e.toString();
      _set(BroadcasterState.error);
    }
  }

  Future<void> _startCapture(StreamModel stream) async {
    final recorder = _recorder ??= AudioRecorder();
    if (!await recorder.hasPermission()) {
      throw Exception('Permission microphone refusée');
    }
    final streamId = stream.id;
    final sessionId = stream.activeSessionId;
    final audioStream = await recorder.startStream(
      const RecordConfig(
        encoder: AudioEncoder.aacLc,
        sampleRate: 44100,
        numChannels: 1,
      ),
    );
    _pushChain = Future<void>.value();
    _captureSub = audioStream.listen((chunk) {
      if (chunk.isNotEmpty) _buffer.add(chunk);
    });
    // Flush accumulated audio on a steady cadence instead of per-frame, to keep
    // the request rate close to the browser's MediaRecorder(500ms) behaviour.
    _flushTimer = Timer.periodic(
      _flushInterval,
      (_) => _flush(streamId, sessionId),
    );
  }

  void _flush(String streamId, String sessionId) {
    if (_buffer.isEmpty) return;
    final bytes = _buffer.takeBytes();
    _pushChain = _pushChain
        .then((_) => _repository.pushChunk(streamId, sessionId, bytes, _mimeType))
        .catchError((Object e) {
          debugPrint('[Broadcaster] push chunk error: $e');
        });
  }

  Future<void> _stopCapture() async {
    _flushTimer?.cancel();
    _flushTimer = null;
    await _captureSub?.cancel();
    _captureSub = null;
    final recorder = _recorder;
    if (recorder != null) {
      try {
        if (await recorder.isRecording()) await recorder.stop();
      } catch (_) {}
    }
    _buffer.clear();
    await _pushChain;
  }

  Future<void> stopStream() async {
    final stream = _currentStream;
    await _stopCapture();
    if (stream != null) {
      try {
        await _repository.stopStream(stream.id, stream.activeSessionId);
      } catch (_) {}
    }
    _currentStream = null;
    _errorMessage = '';
    _set(BroadcasterState.idle);
  }

  Future<void> deleteStream(StreamModel stream) async {
    await _repository.deleteStream(stream.id);
    _ownedStreams = _ownedStreams
        .where((item) => item.id != stream.id)
        .toList();
    notifyListeners();
  }

  void clearError() {
    _errorMessage = '';
    _set(BroadcasterState.idle);
  }

  void _set(BroadcasterState state) {
    _state = state;
    notifyListeners();
  }

  @override
  void dispose() {
    unawaited(_stopCapture());
    _recorder?.dispose();
    super.dispose();
  }
}
