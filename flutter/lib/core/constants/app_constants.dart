/// Application-wide constants for API, endpoints, and configuration.
class AppConstants {
  AppConstants._();

  // API Configuration
  static const String apiBaseUrl = 'http://localhost:8080';
  static const Duration apiTimeout = Duration(seconds: 30);
  static const String bearerTokenPrefix = 'Bearer ';

  // API Endpoints
  static const String authLoginEndpoint = '/api/auth/login';
  static const String authRegisterEndpoint = '/api/auth/register';
  static const String authRefreshEndpoint = '/api/auth/refresh';

  // Streams endpoints
  static const String streamsActiveEndpoint = '/api/streams/active';
  static const String streamsJoinEndpoint = '/api/streams/{id}/join';
  static const String streamDetailEndpoint = '/api/streams/{id}';
  static const String listenHistoryEndpoint = '/api/listen-history';

  // Audio streaming
  static const Duration pollingInterval = Duration(seconds: 5);
  static const String audioStreamMimeType = 'audio/mpeg';

  // Storage keys
  static const String tokenStorageKey = 'auth_token';
  static const String refreshTokenStorageKey = 'refresh_token';
  static const String userIdStorageKey = 'user_id';

  // Error messages
  static const String networkErrorMessage = 'Network error occurred';
  static const String unauthorizedErrorMessage = 'Unauthorized access';
  static const String serverErrorMessage = 'Server error occurred';
  static const String timeoutErrorMessage = 'Request timeout';
}
