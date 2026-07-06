import 'package:flutter/foundation.dart';

import '../api/models/playlist_model.dart';
import '../api/models/track_model.dart';
import '../api/repositories/playlist_repository.dart';

enum LibraryStatus { idle, loading, loaded, error }

class PlaylistNotifier extends ChangeNotifier {
  final PlaylistRepository _repo;

  PlaylistNotifier(this._repo);

  LibraryStatus status = LibraryStatus.idle;
  List<PlaylistModel> playlists = [];
  PlaylistModel? selected;
  TrackModel? lastServedTrack;
  String error = '';

  Future<void> load() async {
    status = LibraryStatus.loading;
    error = '';
    notifyListeners();
    try {
      playlists = await _repo.list();
      status = LibraryStatus.loaded;
    } catch (e) {
      error = e.toString();
      status = LibraryStatus.error;
    }
    notifyListeners();
  }

  Future<void> open(String id) async {
    try {
      selected = await _repo.get(id);
      notifyListeners();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> create(String title) async {
    if (title.trim().isEmpty) return;
    try {
      await _repo.create(title.trim());
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> rename(String id, String title) async {
    if (title.trim().isEmpty) return;
    try {
      await _repo.rename(id, title.trim());
      await load();
      if (selected?.id == id) await open(id);
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> delete(String id) async {
    try {
      await _repo.delete(id);
      if (selected?.id == id) selected = null;
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> addTrack(String playlistId, String trackId) async {
    if (trackId.trim().isEmpty) return;
    try {
      await _repo.addTrack(playlistId, trackId.trim());
      await open(playlistId);
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> removeTrack(String playlistId, String trackId) async {
    try {
      await _repo.removeTrack(playlistId, trackId);
      await open(playlistId);
      await load();
    } catch (e) {
      error = e.toString();
      notifyListeners();
    }
  }

  Future<void> next(String playlistId) async {
    try {
      lastServedTrack = await _repo.next(playlistId);
      await open(playlistId);
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
