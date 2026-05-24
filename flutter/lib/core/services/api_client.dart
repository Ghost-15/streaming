import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../constants/app_constants.dart';

/// Dio HTTP client with JWT interceptor for authenticated requests.
class ApiClient {
  late final Dio _dio;
  final FlutterSecureStorage _secureStorage;

  ApiClient({FlutterSecureStorage? secureStorage})
      : _secureStorage = secureStorage ?? const FlutterSecureStorage() {
    _initializeDio();
  }

  void _initializeDio() {
    _dio = Dio(
      BaseOptions(
        baseUrl: AppConstants.apiBaseUrl,
        connectTimeout: AppConstants.apiTimeout,
        receiveTimeout: AppConstants.apiTimeout,
        contentType: Headers.jsonContentType,
        responseType: ResponseType.json,
      ),
    );

    // Add JWT interceptor for authorization
    _dio.interceptors.add(
      _JwtInterceptor(_secureStorage),
    );

    // Error handling interceptor
    _dio.interceptors.add(
      _ErrorInterceptor(),
    );
  }

  /// GET request
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
  }) async {
    return _dio.get<T>(
      path,
      queryParameters: queryParameters,
    );
  }

  /// POST request
  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
  }) async {
    return _dio.post<T>(
      path,
      data: data,
      queryParameters: queryParameters,
    );
  }

  /// PUT request
  Future<Response<T>> put<T>(
    String path, {
    dynamic data,
  }) async {
    return _dio.put<T>(path, data: data);
  }

  /// DELETE request
  Future<Response<T>> delete<T>(String path) async {
    return _dio.delete<T>(path);
  }

  /// Stream download
  Future<void> download(
    String urlPath,
    String savePath, {
    ProgressCallback? onReceiveProgress,
  }) async {
    await _dio.download(
      urlPath,
      savePath,
      onReceiveProgress: onReceiveProgress,
    );
  }

  void close() {
    _dio.close();
  }
}

/// JWT Bearer Token Interceptor
class _JwtInterceptor extends Interceptor {
  final FlutterSecureStorage _secureStorage;

  _JwtInterceptor(this._secureStorage);

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    try {
      final token =
          await _secureStorage.read(key: AppConstants.tokenStorageKey);

      if (token != null && token.isNotEmpty) {
        options.headers['Authorization'] =
            '${AppConstants.bearerTokenPrefix}$token';
      }

      return handler.next(options);
    } catch (e) {
      return handler.next(options);
    }
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    // Handle 401 Unauthorized - token might be expired
    if (err.response?.statusCode == 401) {
      // TODO: Implement token refresh logic in Sprint 3
    }
    return handler.next(err);
  }
}

/// Error response interceptor for consistent error handling
class _ErrorInterceptor extends Interceptor {
  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    // Log errors or apply custom error handling
    return handler.next(err);
  }
}
