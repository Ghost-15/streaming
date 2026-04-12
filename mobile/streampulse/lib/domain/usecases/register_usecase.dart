// Use Case — Sprint 1 — US-001.
import '../entities/user.dart';
import '../repositories/auth_repository.dart';

class RegisterUseCase {
  final AuthRepository _repo;
  const RegisterUseCase(this._repo);

  Future<User> call({required String email, required String password}) {
    // TODO Sprint 1: password strength validation
    return _repo.register(email: email, password: password);
  }
}
