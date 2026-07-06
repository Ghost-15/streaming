import '../../services/api_service.dart';
import '../models/playlist_model.dart';
import '../models/track_model.dart';

class PlaylistRepository {
  const PlaylistRepository();

  Future<List<PlaylistModel>> list() {
    return ApiService().request(
      uri: 'playlists',
      parser: (res) {
        final list = res is List ? res : (res['data'] as List);
        return list
            .map<PlaylistModel>(
              (e) => PlaylistModel.fromJson(e as Map<String, dynamic>),
            )
            .toList();
      },
    );
  }

  Future<PlaylistModel> get(String id) {
    return ApiService().request(
      uri: 'playlists/$id',
      parser: (res) => PlaylistModel.fromJson(res as Map<String, dynamic>),
    );
  }

  Future<PlaylistModel> create(String title) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'playlists',
      data: {'title': title},
      parser: (res) => PlaylistModel.fromJson(res as Map<String, dynamic>),
    );
  }

  Future<PlaylistModel> rename(String id, String title) {
    return ApiService().request(
      httpMethod: HttpMethod.put,
      uri: 'playlists/$id',
      data: {'title': title},
      parser: (res) => PlaylistModel.fromJson(res as Map<String, dynamic>),
    );
  }

  Future<void> delete(String id) {
    return ApiService().request(
      httpMethod: HttpMethod.delete,
      uri: 'playlists/$id',
    );
  }

  Future<void> addTrack(String playlistId, String trackId) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'playlists/$playlistId/tracks',
      data: {'track_id': trackId},
    );
  }

  Future<void> removeTrack(String playlistId, String trackId) {
    return ApiService().request(
      httpMethod: HttpMethod.delete,
      uri: 'playlists/$playlistId/tracks/$trackId',
    );
  }

  Future<void> reorder(String playlistId, List<String> trackIds) {
    return ApiService().request(
      httpMethod: HttpMethod.put,
      uri: 'playlists/$playlistId/tracks/reorder',
      data: {'track_ids': trackIds},
    );
  }

  Future<TrackModel> next(String playlistId) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'playlists/$playlistId/next',
      parser: (res) => TrackModel.fromJson(res as Map<String, dynamic>),
    );
  }
}
