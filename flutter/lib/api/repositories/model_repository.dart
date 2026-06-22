import '../../services/api_service.dart';

class ModelRepository<T> {
  final String uri;
  final T Function(Map<String, dynamic>) fromJson;

  const ModelRepository({required this.uri, required this.fromJson});

  Future<List<T>> getAll({Map<String, String>? queryParams}) async {
    return ApiService().request(
      uri: uri,
      httpMethod: HttpMethod.get,
      queryParams: queryParams,
      parser: (res) {
        final list = res is List ? res : (res['data'] as List);
        return list.map<T>((e) => fromJson(e as Map<String, dynamic>)).toList();
      },
    );
  }

  Future<T> addOrUpdate({
    required Map<String, dynamic> data,
    String? id,
  }) async {
    return ApiService().request(
      uri: uri,
      httpMethod: id != null ? HttpMethod.put : HttpMethod.post,
      data: data,
      id: id,
      parser: (res) => fromJson(res as Map<String, dynamic>),
    );
  }

  Future<void> delete(String id) {
    return ApiService().request(
      uri: uri,
      httpMethod: HttpMethod.delete,
      id: id,
    );
  }

  Future<T> get(String id) {
    return ApiService().request(
      uri: uri,
      id: id,
      parser: (res) => fromJson(res as Map<String, dynamic>),
    );
  }
}
