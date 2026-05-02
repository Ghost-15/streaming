// Repository interface — domain layer.
import '../entities/playlist.dart';

abstract interface class PlaylistRepository {
  Future<List<Playlist>> listByOwner(String ownerId);
  Future<Playlist> create({required String title});
  Future<void> delete(String playlistId);
  Future<void> addTrack({required String playlistId, required String trackId});
  Future<void> removeTrack({required String playlistId, required String trackId});
}
