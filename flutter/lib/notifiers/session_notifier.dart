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
    final savedToken = await StorageService.get(StorageKey.token);
    final savedUser = await StorageService.get(StorageKey.userData);
    if (savedToken == null || savedUser == null) return;

    try {
      token = savedToken;
      user = UserModel.fromJson(jsonDecode(savedUser) as Map<String, dynamic>);
      notifyListeners();
    } catch (_) {
      await _clearStorage();
    }
  }

  void onAuthentication(AuthResponse response) {
    user = response.user;
    token = response.token;
    StorageService.save(StorageKey.token, response.token);
    StorageService.save(StorageKey.userData, jsonEncode(response.user.toJson()));
    notifyListeners();
  }

  void logout() {
    user = null;
    token = null;
    _clearStorage();
    notifyListeners();
  }

  bool get isAuthenticated => user != null && token != null;

  Future<void> _clearStorage() async {
    await StorageService.remove(StorageKey.token);
    await StorageService.remove(StorageKey.userData);
  }
}
