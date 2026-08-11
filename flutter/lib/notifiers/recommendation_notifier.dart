import 'package:flutter/foundation.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

class RecommendationNotifier extends ChangeNotifier {
  final StreamRepository _repository;

  RecommendationNotifier(this._repository);

  List<StreamModel> _recommendations = [];
  bool _isLoading = false;
  bool _loaded = false;

  List<StreamModel> get recommendations => _recommendations;
  bool get isLoading => _isLoading;
  bool get hasRecommendations => _recommendations.isNotEmpty;

  Future<void> load() async {
    if (_isLoading) return;
    _isLoading = true;
    notifyListeners();
    try {
      _recommendations = await _repository.getRecommendations();
      _loaded = true;
    } catch (e) {
      debugPrint('[Reco] load failed: $e');
      _recommendations = [];
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> reload() async {
    _loaded = false;
    await load();
  }

  bool get isLoaded => _loaded;
}
