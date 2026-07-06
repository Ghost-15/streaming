import 'package:flutter/foundation.dart';

import '../api/models/track_model.dart';
import '../api/repositories/favorite_repository.dart';
import 'playlist_notifier.dart';

class FavoriteNotifier extends ChangeNotifier {
  final FavoriteRepository _repo;

  FavoriteNotifier(this._repo);

  LibraryStatus status = LibraryStatus.idle;
  List<TrackModel> tracks = [];
  String error = '';

  Future<void> load() async {
    status = LibraryStatus.loading;
    error = '';
    notifyListeners();
    try {
      tracks = await _repo.list();
      status = LibraryStatus.loaded;
    } catch (e) {
      error = e.toString();
      status = LibraryStatus.error;
    }
    notifyListeners();
  }

  Future<void> add(String trackId) async {
    if (trackId.trim().isEmpty) return;
    try {
      await _repo.add(trackId.trim());
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> remove(String trackId) async {
    try {
      await _repo.remove(trackId);
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  void clearError() {
    error = '';
    notifyListeners();
  }
}
