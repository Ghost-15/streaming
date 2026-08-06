import 'package:flutter/foundation.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum BroadcasterState { idle, loading, streaming, error }

// Non-Web fallback used by VM tests and unsupported desktop/mobile targets.
// The production Web export provides microphone capture via MediaRecorder.
class BroadcasterNotifier extends ChangeNotifier {
  BroadcasterNotifier(this._repository);

  final StreamRepository _repository;
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
    } else {
      _errorMessage =
          'La capture audio en direct est disponible dans le client Web.';
    }
    _set(BroadcasterState.error);
  }

  Future<void> restartStream(StreamModel stream) async {
    _errorMessage =
        'La capture audio en direct est disponible dans le client Web.';
    _set(BroadcasterState.error);
  }

  Future<void> stopStream() async {
    final stream = _currentStream;
    if (stream != null) {
      await _repository.stopStream(stream.id, stream.activeSessionId);
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
}
