import 'dart:async';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

enum StorageKey { token, refreshToken, userId, userData }

class StorageService {
  static const _storage = FlutterSecureStorage();

  // flutter_secure_storage_web lazily creates a shared AES-GCM key. If two
  // first writes race, each can encrypt with a different key and one value
  // becomes unreadable. Keep every storage operation in invocation order so
  // writes, reads, and logout cleanup cannot overtake each other.
  static Future<void> _operationTail = Future<void>.value();

  static Future<T> _serialized<T>(Future<T> Function() operation) {
    final result = Completer<T>();
    _operationTail = _operationTail.then((_) async {
      try {
        result.complete(await operation());
      } catch (error, stackTrace) {
        result.completeError(error, stackTrace);
      }
    });
    return result.future;
  }

  static Future<void> save(StorageKey key, String value) => _serialized(
    () => _storage.write(key: key.name, value: value),
  );

  static Future<String?> get(StorageKey key) => _serialized(() async {
    try {
      return await _storage.read(key: key.name);
    } catch (error) {
      // AES-GCM authentication failure means this single ciphertext cannot be
      // recovered. Remove it so the app can return to login instead of leaking
      // a raw browser OperationError through every authenticated request.
      if (!_isCryptoOperationError(error)) rethrow;
      await _storage.delete(key: key.name);
      return null;
    }
  });

  static Future<void> remove(StorageKey key) => _serialized(
    () => _storage.delete(key: key.name),
  );

  static bool _isCryptoOperationError(Object error) =>
      error.toString().contains('OperationError');
}
