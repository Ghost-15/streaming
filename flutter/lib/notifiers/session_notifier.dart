import 'dart:convert';

import 'package:flutter/foundation.dart';

import '../api/models/user_model.dart';
import '../api/repositories/auth_repository.dart';
import '../services/storage_service.dart';

class SessionNotifier extends ChangeNotifier {
  UserModel? user;
  String? token;

  // Restore session from storage on app start.
  // Called once in main() before runApp so the router has the right state.
  Future<void> init() async {
    try {
      final savedToken = await StorageService.get(StorageKey.token);
      final savedUser = await StorageService.get(StorageKey.userData);
      if (savedToken == null || savedUser == null) {
        // Remove the remaining half of a partial/corrupted session.
        if (savedToken != null || savedUser != null) await _clearStorage();
        return;
      }

      token = savedToken;
      user = UserModel.fromJson(jsonDecode(savedUser) as Map<String, dynamic>);
      notifyListeners();
    } catch (_) {
      // Storage unavailable (IndexedDB blocked, private mode, corrupted key).
      // Start unauthenticated rather than crashing.
      try {
        await _clearStorage();
      } catch (_) {}
    }
  }

  Future<void> onAuthentication(AuthResponse response) async {
    try {
      await StorageService.save(StorageKey.token, response.token);
      await StorageService.save(
        StorageKey.userData,
        jsonEncode(response.user.toJson()),
      );
    } catch (_) {
      try {
        await _clearStorage();
      } catch (_) {}
      rethrow;
    }

    user = response.user;
    token = response.token;
    notifyListeners();
  }

  Future<void> logout() async {
    user = null;
    token = null;
    notifyListeners();
    try {
      await _clearStorage();
    } catch (_) {}
  }

  bool get isAuthenticated => user != null && token != null;

  Future<void> _clearStorage() async {
    await StorageService.remove(StorageKey.token);
    await StorageService.remove(StorageKey.userData);
  }
}
