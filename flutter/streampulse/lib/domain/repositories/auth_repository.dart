// Repository interface — domain layer.
// Implemented in data/repositories/auth_repository_impl.dart
import '../entities/user.dart';

abstract interface class AuthRepository {
  Future<User> register({required String email, required String password});
  Future<String> login({required String email, required String password});
  Future<void> logout();
  Future<User?> currentUser();
}
