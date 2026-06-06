import 'package:flutter_secure_storage/flutter_secure_storage.dart';

enum StorageKey { token, refreshToken, userId }

class StorageService {
  static const _storage = FlutterSecureStorage();

  static Future<void> save(StorageKey key, String value) async {
    await _storage.write(key: key.name, value: value);
  }

  static Future<String?> get(StorageKey key) async {
    return _storage.read(key: key.name);
  }

  static Future<void> remove(StorageKey key) async {
    await _storage.delete(key: key.name);
  }
}
