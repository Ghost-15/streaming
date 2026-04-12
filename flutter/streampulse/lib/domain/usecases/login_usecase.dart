// Use Case — orchestrates domain logic, depends only on repository interfaces.
// Sprint 1 — US-001.
import '../repositories/auth_repository.dart';

class LoginUseCase {
  final AuthRepository _repo;
  const LoginUseCase(this._repo);

  /// Returns the JWT token on success.
  Future<String> call({required String email, required String password}) async {
    // TODO Sprint 1: validate email format + call repo
    return _repo.login(email: email, password: password);
  }
}
