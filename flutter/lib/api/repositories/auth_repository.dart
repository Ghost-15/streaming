import '../../services/api_service.dart';
import '../models/user_model.dart';

class AuthResponse {
  final String token;
  final UserModel user;

  AuthResponse({required this.token, required this.user});

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    return AuthResponse(
      token: json['token'],
      user: UserModel.fromJson(json['user']),
    );
  }

  @override
  String toString() => token;
}

class AuthRepository {
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

  Future<AuthResponse> refreshUser() {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/refresh',
      parser: (res) => AuthResponse.fromJson(res),
    );
  }
}
