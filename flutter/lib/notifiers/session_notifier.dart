import 'dart:convert';

import 'package:flutter/foundation.dart';

import '../api/models/user_model.dart';
import '../api/repositories/auth_repository.dart';
import '../services/api_service.dart';
import '../services/storage_service.dart';

class SessionNotifier extends ChangeNotifier {
  final AuthRepository _repo;

  SessionNotifier([this._repo = const AuthRepository()]) {
    // ApiService replays a request once after a successful renewal.
    ApiService.onRefreshNeeded = tryRefresh;
  }

  UserModel? user;
  String? token;
  String? refreshToken;

  // Concurrent 401s must trigger a single renewal, not one per request.
  Future<bool>? _inFlightRefresh;

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
      refreshToken = await StorageService.get(StorageKey.refreshToken);
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
        StorageKey.refreshToken,
        response.refreshToken,
      );
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
    refreshToken = response.refreshToken;
    notifyListeners();
  }

  // Renews the access token from the stored refresh token. Returns false when
  // no renewal is possible, which leaves the caller to surface the 401.
  Future<bool> tryRefresh() {
    return _inFlightRefresh ??= _refresh().whenComplete(() {
      _inFlightRefresh = null;
    });
  }

  Future<bool> _refresh() async {
    final current = refreshToken;
    if (current == null || current.isEmpty) return false;
    try {
      final response = await _repo.refresh(current);
      await onAuthentication(response);
      return true;
    } catch (_) {
      // Refresh token expired, revoked or replayed — the session is over.
      return false;
    }
  }

  Future<void> logout() async {
    final current = refreshToken;

    user = null;
    token = null;
    refreshToken = null;
    notifyListeners();

    if (current != null && current.isNotEmpty) {
      try {
        await _repo.revoke(current);
      } catch (_) {
        // Best effort: the local session is cleared regardless.
      }
    }
    try {
      await _clearStorage();
    } catch (_) {}
  }

  bool get isAuthenticated => user != null && token != null;

  Future<void> _clearStorage() async {
    await StorageService.remove(StorageKey.token);
    await StorageService.remove(StorageKey.refreshToken);
    await StorageService.remove(StorageKey.userData);
  }
}
