import '../../services/api_service.dart';
import '../models/track_model.dart';

class FavoriteRepository {
  const FavoriteRepository();

  Future<List<TrackModel>> list() {
    return ApiService().request(
      uri: 'favorites',
      notifyOnUnauthorized: false,
      parser: (res) {
        final list = res is List ? res : (res['data'] as List);
        return list
            .map<TrackModel>(
              (e) => TrackModel.fromJson(e as Map<String, dynamic>),
            )
            .toList();
      },
    );
  }

  Future<void> add(String trackId) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'favorites',
      data: {'track_id': trackId},
    );
  }

  Future<void> remove(String trackId) {
    return ApiService().request(
      httpMethod: HttpMethod.delete,
      uri: 'favorites/$trackId',
    );
  }
}
