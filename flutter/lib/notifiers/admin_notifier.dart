import 'package:flutter/foundation.dart';

import '../api/models/admin_stats_model.dart';
import '../api/models/role.dart';
import '../api/models/user_model.dart';
import '../api/repositories/admin_repository.dart';

enum AdminStatus { idle, loading, loaded, error }

class AdminNotifier extends ChangeNotifier {
  final AdminRepository _repo;

  AdminNotifier(this._repo);

  AdminStatus _usersStatus = AdminStatus.idle;
  AdminStatus _statsStatus = AdminStatus.idle;
  List<UserModel> _users = [];
  int _total = 0;
  AdminStats? _stats;
  String _error = '';

  AdminStatus get usersStatus => _usersStatus;
  AdminStatus get statsStatus => _statsStatus;
  List<UserModel> get users => _users;
  int get total => _total;
  AdminStats? get stats => _stats;
  String get error => _error;

  Future<void> loadUsers({int page = 1}) async {
    _usersStatus = AdminStatus.loading;
    _error = '';
    notifyListeners();
    try {
      final result = await _repo.listUsers(page: page);
      _users = result.users;
      _total = result.total;
      _usersStatus = AdminStatus.loaded;
    } catch (e) {
      _error = e.toString();
      _usersStatus = AdminStatus.error;
    }
    notifyListeners();
  }

  Future<void> loadStats() async {
    _statsStatus = AdminStatus.loading;
    notifyListeners();
    try {
      _stats = await _repo.getStats();
      _statsStatus = AdminStatus.loaded;
    } catch (e) {
      _error = e.toString();
      _statsStatus = AdminStatus.error;
    }
    notifyListeners();
  }

  Future<void> updateRole(String userId, Role role) async {
    try {
      await _repo.updateRole(userId, role.apiValue);
      _users = _users.map((u) {
        return u.id == userId
            ? UserModel(
                id: u.id,
                email: u.email,
                firstName: u.firstName,
                lastName: u.lastName,
                role: role,
                suspendedAt: u.suspendedAt,
              )
            : u;
      }).toList();
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
    }
  }

  Future<void> suspendUser(String userId, {required bool suspend}) async {
    try {
      await _repo.suspendUser(userId, suspend: suspend);
      await loadUsers();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
    }
  }

  void clearError() {
    _error = '';
    notifyListeners();
  }
}
