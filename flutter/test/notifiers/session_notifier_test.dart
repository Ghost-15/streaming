import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/api/models/role.dart';
import 'package:streampulse/api/models/user_model.dart';
import 'package:streampulse/api/repositories/auth_repository.dart';
import 'package:streampulse/notifiers/session_notifier.dart';
import 'package:streampulse/services/storage_service.dart';

const _secureStorageChannel = MethodChannel(
  'plugins.it_nomads.com/flutter_secure_storage',
);

void _mockSecureStorage([Future<Object?> Function(MethodCall call)? handler]) {
  TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
      .setMockMethodCallHandler(
        _secureStorageChannel,
        handler ?? (call) async => null,
      );
}

AuthResponse _fakeResponse(Role role) => AuthResponse(
  token: 'tok-123',
  refreshToken: 'refresh-123',
  expiresIn: 3600,
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

    test('onAuthentication sets user and token', () async {
      await notifier.onAuthentication(_fakeResponse(Role.user));
      expect(notifier.isAuthenticated, isTrue);
      expect(notifier.user?.email, 'test@test.com');
      expect(notifier.token, 'tok-123');
    });

    test('onAuthentication notifies listeners', () async {
      bool notified = false;
      notifier.addListener(() => notified = true);
      await notifier.onAuthentication(_fakeResponse(Role.user));
      expect(notified, isTrue);
    });

    test('logout clears user and token', () async {
      await notifier.onAuthentication(_fakeResponse(Role.admin));
      await notifier.logout();
      expect(notifier.isAuthenticated, isFalse);
      expect(notifier.user, isNull);
      expect(notifier.token, isNull);
    });

    test('logout notifies listeners', () async {
      await notifier.onAuthentication(_fakeResponse(Role.user));
      bool notified = false;
      notifier.addListener(() => notified = true);
      await notifier.logout();
      expect(notified, isTrue);
    });

    test('role is preserved after authentication', () async {
      await notifier.onAuthentication(_fakeResponse(Role.diffuseur));
      expect(notifier.user?.role, Role.diffuseur);
    });

    test('secure storage operations never overlap', () async {
      var activeCalls = 0;
      var maxActiveCalls = 0;
      final savedKeys = <String>[];
      _mockSecureStorage((call) async {
        if (call.method != 'write') return null;
        activeCalls++;
        if (activeCalls > maxActiveCalls) maxActiveCalls = activeCalls;
        await Future<void>.delayed(const Duration(milliseconds: 10));
        final arguments = call.arguments as Map<Object?, Object?>;
        savedKeys.add(arguments['key']! as String);
        activeCalls--;
        return null;
      });

      await Future.wait([
        StorageService.save(StorageKey.token, 'token'),
        StorageService.save(StorageKey.userData, 'user'),
      ]);

      expect(maxActiveCalls, 1);
      expect(savedKeys, ['token', 'userData']);
    });

    test(
      'corrupted encrypted value is removed and treated as missing',
      () async {
        String? deletedKey;
        _mockSecureStorage((call) async {
          final arguments = call.arguments as Map<Object?, Object?>;
          if (call.method == 'read') {
          throw PlatformException(code: 'OperationError');
          }
          if (call.method == 'delete') {
            deletedKey = arguments['key']! as String;
          }
          return null;
        });

        final value = await StorageService.get(StorageKey.token);

        expect(value, isNull);
        expect(deletedKey, 'token');
      },
    );
  });
}
