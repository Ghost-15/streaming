import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/api/models/admin_stats_model.dart';
import 'package:streampulse/api/models/role.dart';
import 'package:streampulse/api/models/user_model.dart';
import 'package:streampulse/api/repositories/admin_repository.dart';
import 'package:streampulse/notifiers/admin_notifier.dart';

// ── Fake ─────────────────────────────────────────────────────────────────────

UserModel _user(String id, {Role role = Role.user}) => UserModel(
  id: id,
  email: '$id@test.com',
  firstName: 'First',
  lastName: 'Last',
  role: role,
);

class _FakeAdminRepo extends AdminRepository {
  List<UserModel> users;
  AdminStats? stats;
  bool shouldThrow;

  _FakeAdminRepo({this.users = const [], this.stats, this.shouldThrow = false});

  @override
  Future<({List<UserModel> users, int total})> listUsers({
    int page = 1,
    int limit = 20,
  }) async {
    if (shouldThrow) throw Exception('network error');
    return (users: users, total: users.length);
  }

  @override
  Future<void> updateRole(String userId, String role) async {
    if (shouldThrow) throw Exception('network error');
  }

  @override
  Future<void> suspendUser(String userId, {required bool suspend}) async {
    if (shouldThrow) throw Exception('not implemented');
  }

  @override
  Future<AdminStats> getStats() async {
    if (shouldThrow) throw Exception('network error');
    return stats ??
        const AdminStats(totalUsers: 3, byRole: {'user': 2, 'admin': 1});
  }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('AdminNotifier', () {
    test('initial state is idle', () {
      final n = AdminNotifier(_FakeAdminRepo());
      expect(n.usersStatus, AdminStatus.idle);
      expect(n.statsStatus, AdminStatus.idle);
      expect(n.users, isEmpty);
      expect(n.stats, isNull);
      expect(n.error, isEmpty);
      n.dispose();
    });

    test('loadUsers() transitions to loaded', () async {
      final repo = _FakeAdminRepo(users: [_user('u-1'), _user('u-2')]);
      final n = AdminNotifier(repo);
      await n.loadUsers();
      expect(n.usersStatus, AdminStatus.loaded);
      expect(n.users.length, 2);
      expect(n.total, 2);
      n.dispose();
    });

    test('loadUsers() on error sets error state', () async {
      final n = AdminNotifier(_FakeAdminRepo(shouldThrow: true));
      await n.loadUsers();
      expect(n.usersStatus, AdminStatus.error);
      expect(n.error, isNotEmpty);
      n.dispose();
    });

    test('loadUsers() notifies listeners', () async {
      final n = AdminNotifier(_FakeAdminRepo());
      int calls = 0;
      n.addListener(() => calls++);
      await n.loadUsers();
      expect(calls, greaterThan(0));
      n.dispose();
    });

    test('loadStats() sets stats', () async {
      final repo = _FakeAdminRepo(
        stats: const AdminStats(
          totalUsers: 10,
          byRole: {'user': 8, 'admin': 2},
        ),
      );
      final n = AdminNotifier(repo);
      await n.loadStats();
      expect(n.statsStatus, AdminStatus.loaded);
      expect(n.stats?.totalUsers, 10);
      expect(n.stats?.byRole['admin'], 2);
      n.dispose();
    });

    test('updateRole() updates user role in list', () async {
      final repo = _FakeAdminRepo(users: [_user('u-1', role: Role.user)]);
      final n = AdminNotifier(repo);
      await n.loadUsers();
      await n.updateRole('u-1', Role.diffuseur);
      expect(n.users.first.role, Role.diffuseur);
      n.dispose();
    });

    test('clearError() resets error and notifies', () async {
      final n = AdminNotifier(_FakeAdminRepo(shouldThrow: true));
      await n.loadUsers();
      bool notified = false;
      n.addListener(() => notified = true);
      n.clearError();
      expect(n.error, isEmpty);
      expect(notified, isTrue);
      n.dispose();
    });
  });
}
