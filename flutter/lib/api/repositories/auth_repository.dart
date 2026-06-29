import 'dart:convert';

import '../../services/api_service.dart';
import '../models/role.dart';
import '../models/user_model.dart';

class AuthResponse {
  final String token;
  final UserModel user;

  AuthResponse({required this.token, required this.user});

  @override
  String toString() => token;
}

class AuthRepository {
  /// POST /auth/login returns a JWT token only.
  /// The backend only returns a token, so we rebuild the user from the JWT claims
  /// (sub = id, email, role) instead of expecting a "user" object.
  Future<AuthResponse> authenticate(Map<String, dynamic> data) async {
    final token = await ApiService().request<String>(
      httpMethod: HttpMethod.post,
      uri: 'auth/login',
      data: data,
      parser: (res) => res['token'] as String,
    );
    return AuthResponse(token: token, user: _userFromJwt(token));
  }

  /// POST /auth/register → { "id", "email" }, then auto-login to obtain a token.
  Future<AuthResponse> register(Map<String, dynamic> data) async {
    await ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'auth/register',
      data: data,
      parser: (res) => res,
    );
    return authenticate(data);
  }

  /// Decodes the JWT payload (no signature check — display only) into a UserModel.
  UserModel _userFromJwt(String token) {
    try {
      final parts = token.split('.');
      if (parts.length != 3) {
        throw const FormatException('malformed jwt');
      }
      final payload = utf8.decode(
        base64Url.decode(base64Url.normalize(parts[1])),
      );
      final claims = jsonDecode(payload) as Map<String, dynamic>;
      return UserModel.fromJson({
        'id': claims['sub'] ?? '',
        'email': claims['email'] ?? '',
        'role': claims['role'] ?? 'user',
      });
    } catch (_) {
      return const UserModel(
        id: '',
        email: '',
        firstName: '',
        lastName: '',
        role: Role.anon,
      );
    }
  }
}
