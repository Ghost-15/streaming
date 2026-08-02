import 'package:flutter/foundation.dart';

import '../../services/api_service.dart';
import '../../config/app_config.dart';
import '../models/stream_model.dart';
import 'model_repository.dart';

class StreamRepository extends ModelRepository<StreamModel> {
  const StreamRepository()
    : super(uri: 'streams', fromJson: StreamModel.fromJson);

  Future<List<StreamModel>> getActive() {
    return ApiService().request(
      uri: 'streams',
      parser: (res) {
        final list = res is List ? res : (res['data'] as List);
        return list.map<StreamModel>((e) {
          final stream = StreamModel.fromJson(e as Map<String, dynamic>);
          return stream.copyWith(
            streamUrl: '${AppConfig.apiBaseUrl}/streams/${stream.id}/audio',
          );
        }).toList();
      },
    );
  }

  Future<void> joinStream(String id) {
    return ApiService().request<dynamic>(
      httpMethod: HttpMethod.post,
      uri: 'streams/$id/listen',
      notifyOnUnauthorized: false,
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

  Future<void> stopStream(String id) {
    return ApiService().request(
      httpMethod: HttpMethod.put,
      uri: 'streams/$id/stop',
    );
  }

  Future<void> pushChunk(String id, Uint8List chunk, String contentType) {
    return ApiService().rawPost(
      uri: 'streams/$id/push',
      body: chunk,
      contentType: contentType,
    );
  }
}
