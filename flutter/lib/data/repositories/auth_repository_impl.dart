// Data layer — implements domain/repositories/auth_repository.dart
// Sprint 1 — US-001.
import '../../domain/entities/user.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/api_client.dart';

class AuthRepositoryImpl implements AuthRepository {
  // ignore: unused_field — sera utilisé Sprint 1 (US-001)
  final ApiClient _client;
  const AuthRepositoryImpl(this._client);

  @override
  Future<User> register({required String email, required String password}) async {
    // TODO Sprint 1: POST /api/v1/auth/register
    // Parse response → User entity
    throw UnimplementedError('register not yet implemented');
  }

  @override
  Future<String> login({required String email, required String password}) async {
    // TODO Sprint 1: POST /api/v1/auth/login → returns JWT token
    throw UnimplementedError('login not yet implemented');
  }

  @override
  Future<void> logout() async {
    // TODO Sprint 1: clear stored JWT from flutter_secure_storage
  }

  @override
  Future<User?> currentUser() async {
    // TODO Sprint 1: decode stored JWT, return User or null if expired
    return null;
  }
}
