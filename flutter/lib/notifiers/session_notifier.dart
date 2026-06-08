import 'package:flutter/foundation.dart';

import '../api/models/user_model.dart';
import '../api/repositories/auth_repository.dart';
import '../services/storage_service.dart';

class SessionNotifier extends ChangeNotifier {
  UserModel? user;
  String? token;

  void onAuthentication(AuthResponse response) {
    user = response.user;
    token = response.token;
    StorageService.save(StorageKey.token, response.token);
    notifyListeners();
  }

  void logout() {
    user = null;
    token = null;
    StorageService.remove(StorageKey.token);
    notifyListeners();
  }

  bool get isAuthenticated => user != null && token != null;
}
