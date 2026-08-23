import '../../services/api_service.dart';
import '../models/user_model.dart';

class AuthResponse {
  final String token;
  final String refreshToken;
  final int expiresIn;
  final UserModel user;

  AuthResponse({
    required this.token,
    required this.refreshToken,
    required this.expiresIn,
    required this.user,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      token: json['token'] ?? '',
      refreshToken: json['refresh_token'] ?? '',
      expiresIn: json['expires_in'] as int? ?? 0,
      user: UserModel.fromJson(json['user'] ?? json),
    );
  }

  @override
  String toString() => token;
}

class AuthRepository {
  const AuthRepository();

  Future<AuthResponse> authenticate(Map<String, dynamic> data) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/login',
      data: data,
      parser: (res) => AuthResponse.fromJson(res),
    );
  }

  Future<AuthResponse> register(Map<String, dynamic> data) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/register',
      data: data,
      parser: (res) => AuthResponse.fromJson(res),
    );
  }

  // Exchanges a refresh token for a new pair. The server rotates the refresh
  // token, so the returned value replaces the one that was sent.
  Future<AuthResponse> refresh(String refreshToken) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/refresh',
      data: {'refresh_token': refreshToken},
      notifyOnUnauthorized: false,
      allowRefreshRetry: false,
      parser: (res) => AuthResponse.fromJson(res),
    );
  }

  // Revokes the refresh token server side so a stolen value cannot be reused.
  Future<void> revoke(String refreshToken) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/logout',
      data: {'refresh_token': refreshToken},
      notifyOnUnauthorized: false,
      allowRefreshRetry: false,
    );
  }

  Future<UserModel> me() {
    return ApiService().request(
      uri: 'auth/me',
      parser: (res) => UserModel.fromJson(res['user'] ?? res),
    );
  }
}
