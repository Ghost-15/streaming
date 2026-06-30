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
      uri: 'streams',
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

  /// Starts a stream. stream_url is optional — the backend falls back to a demo
  /// audio source when omitted, so the stream is immediately playable.
  Future<StreamModel> startStream(String title, {String? streamUrl}) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'streams',
      data: {
        'title': title,
        if (streamUrl != null && streamUrl.isNotEmpty) 'stream_url': streamUrl,
      },
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
