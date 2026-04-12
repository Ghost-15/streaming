import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/domain/usecases/login_usecase.dart';

import '../../mocks/mock_auth_repository.dart';

void main() {
  group('LoginUseCase', () {
    late MockAuthRepository mockRepo;
    late LoginUseCase useCase;

    setUp(() {
      mockRepo = MockAuthRepository();
      useCase = LoginUseCase(mockRepo);
    });

    test('returns token on success', () async {
      mockRepo.loginResult = 'fake-jwt-token';

      final token = await useCase(
        email: 'test@example.com',
        password: 'password123',
      );

      expect(token, equals('fake-jwt-token'));
    });

    test('throws on error', () async {
      mockRepo.shouldThrow = true;

      expect(
        () => useCase(email: 'bad@example.com', password: 'wrong'),
        throwsException,
      );
    });
  });
}
