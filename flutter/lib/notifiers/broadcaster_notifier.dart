import 'package:flutter/foundation.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum BroadcasterState { idle, loading, streaming, error }

class BroadcasterNotifier extends ChangeNotifier {
  final StreamRepository _repository;

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
      _set(BroadcasterState.streaming);
    } catch (e) {
      _errorMessage = e.toString();
      _set(BroadcasterState.error);
    }
  }

  Future<void> stopStream() async {
    if (_currentStream == null) return;
    _set(BroadcasterState.loading);
    try {
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
}
