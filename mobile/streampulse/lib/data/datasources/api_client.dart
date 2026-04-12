// Data source — HTTP client wrapping Dio.
// All calls target the Go Gin API.
// Sprint 1 — US-001.
import 'package:dio/dio.dart';

class ApiClient {
  late final Dio _dio;

  ApiClient({required String baseUrl, String? token}) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {
        'Content-Type': 'application/json',
        if (token != null) 'Authorization': 'Bearer $token',
      },
    ));

    // TODO Sprint 2 — US-008: add OTEL interceptor for distributed tracing
    // TODO Sprint 1: add auth interceptor to auto-refresh JWT
  }

  Future<Response> get(String path, {Map<String, dynamic>? queryParams}) =>
      _dio.get(path, queryParameters: queryParams);

  Future<Response> post(String path, {Object? data}) =>
      _dio.post(path, data: data);

  Future<Response> put(String path, {Object? data}) =>
      _dio.put(path, data: data);

  Future<Response> delete(String path) => _dio.delete(path);
}
