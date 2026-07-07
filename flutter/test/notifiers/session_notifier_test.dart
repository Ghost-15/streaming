import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/api/models/role.dart';
import 'package:streampulse/api/models/user_model.dart';
import 'package:streampulse/api/repositories/auth_repository.dart';
import 'package:streampulse/notifiers/session_notifier.dart';

const _secureStorageChannel =
    MethodChannel('plugins.it_nomads.com/flutter_secure_storage');

void _mockSecureStorage() {
  TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
      .setMockMethodCallHandler(_secureStorageChannel, (call) async => null);
}

AuthResponse _fakeResponse(Role role) => AuthResponse(
      token: 'tok-123',
      user: UserModel(
        id: 'u-1',
        email: 'test@test.com',
        firstName: 'Test',
        lastName: 'User',
        role: role,
      ),
    );

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('SessionNotifier', () {
    late SessionNotifier notifier;

    setUp(() {
      _mockSecureStorage();
      notifier = SessionNotifier();
    });
    tearDown(() => notifier.dispose());

    test('initial state: not authenticated', () {
      expect(notifier.isAuthenticated, isFalse);
      expect(notifier.user, isNull);
      expect(notifier.token, isNull);
    });

    test('onAuthentication sets user and token', () {
      notifier.onAuthentication(_fakeResponse(Role.user));
      expect(notifier.isAuthenticated, isTrue);
      expect(notifier.user?.email, 'test@test.com');
      expect(notifier.token, 'tok-123');
    });

    test('onAuthentication notifies listeners', () {
      bool notified = false;
      notifier.addListener(() => notified = true);
      notifier.onAuthentication(_fakeResponse(Role.user));
      expect(notified, isTrue);
    });

    test('logout clears user and token', () {
      notifier.onAuthentication(_fakeResponse(Role.admin));
      notifier.logout();
      expect(notifier.isAuthenticated, isFalse);
      expect(notifier.user, isNull);
      expect(notifier.token, isNull);
    });

    test('logout notifies listeners', () {
      notifier.onAuthentication(_fakeResponse(Role.user));
      bool notified = false;
      notifier.addListener(() => notified = true);
      notifier.logout();
      expect(notified, isTrue);
    });

    test('role is preserved after authentication', () {
      notifier.onAuthentication(_fakeResponse(Role.diffuseur));
      expect(notifier.user?.role, Role.diffuseur);
    });
  });
}
