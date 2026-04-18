import 'package:streampulse/domain/entities/user.dart';
import 'package:streampulse/domain/repositories/auth_repository.dart';

class MockAuthRepository implements AuthRepository {
  String? loginResult;
  bool shouldThrow = false;

  @override
  Future<String> login({required String email, required String password}) async {
    if (shouldThrow) throw Exception('login failed');
    return loginResult ?? 'mock-token';
  }

  @override
  Future<User> register({required String email, required String password}) async {
    if (shouldThrow) throw Exception('register failed');
    return const User(
      id: 'mock-id',
      email: 'test@test.com',
      firstName: 'Test',
      lastName: 'User',
      role: UserRole.user,
    );
  }

  @override
  Future<void> logout() async {}

  @override
  Future<User?> currentUser() async => null;
}
