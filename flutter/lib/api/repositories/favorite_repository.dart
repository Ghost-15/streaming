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

  Future<void> add(String streamId) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'favorites',
      data: {'stream_id': streamId},
    );
  }

  Future<void> remove(String streamId) {
    return ApiService().request(
      httpMethod: HttpMethod.delete,
      uri: 'favorites/$streamId',
    );
  }
}
