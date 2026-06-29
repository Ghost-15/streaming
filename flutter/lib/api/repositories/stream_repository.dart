import '../../services/api_service.dart';
import '../models/stream_model.dart';
import 'model_repository.dart';

class StreamRepository extends ModelRepository<StreamModel> {
  const StreamRepository()
      : super(
          uri: 'streams',
          fromJson: StreamModel.fromJson,
        );

  Future<List<StreamModel>> getActive() {
    return ApiService().request(
      uri: 'streams/active',
      parser: (res) {
        final list = res is List ? res : (res['data'] as List);
        return list
            .map<StreamModel>((e) => StreamModel.fromJson(e as Map<String, dynamic>))
            .toList();
      },
    );
  }

  Future<StreamModel> joinStream(String id) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'streams/$id/join',
      parser: (res) => StreamModel.fromJson(res),
    );
  }

  Future<StreamModel> startStream(String title) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'streams',
      data: {'title': title},
      parser: (res) => StreamModel.fromJson(res),
    );
  }

  Future<void> stopStream(String id) async {
    await ApiService().request<void>(
      httpMethod: HttpMethod.put,
      uri: 'streams/$id/stop',
    );
  }
}
