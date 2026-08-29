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

  Future<void> add(String streamId) async {
    if (streamId.trim().isEmpty) return;
    try {
      await _repo.add(streamId.trim());
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> remove(String streamId) async {
    try {
      await _repo.remove(streamId);
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
