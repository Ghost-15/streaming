import 'package:flutter/foundation.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

class StreamNotifier extends ChangeNotifier {
  List<StreamModel> streams = [];
  bool isLoading = false;
  String? error;

  final StreamRepository _repository;

  StreamNotifier(this._repository);

  Future<void> loadActive() async {
    isLoading = true;
    error = null;
    notifyListeners();

    try {
      streams = await _repository.getActive();
    } catch (e) {
      error = e.toString();
    } finally {
      isLoading = false;
      notifyListeners();
    }
  }
}
