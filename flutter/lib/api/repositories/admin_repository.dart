import '../../services/api_service.dart';
import '../models/admin_stats_model.dart';
import '../models/user_model.dart';

class AdminRepository {
  const AdminRepository();

  Future<({List<UserModel> users, int total})> listUsers({
    int page = 1,
    int limit = 20,
  }) {
    return ApiService().request(
      uri: 'admin/users',
      queryParams: {'page': '$page', 'limit': '$limit'},
      parser: (res) {
        final list = (res['data'] ?? res['users']) as List;
        return (
          users: list
              .map((e) => UserModel.fromJson(e as Map<String, dynamic>))
              .toList(),
          total: res['total'] as int? ?? 0,
        );
      },
    );
  }

  Future<void> updateRole(String userId, String role) {
    return ApiService().request(
      httpMethod: HttpMethod.put,
      uri: 'admin/users/$userId/role',
      data: {'role': role},
    );
  }

  Future<void> suspendUser(String userId, {required bool suspend}) {
    return ApiService().request(
      httpMethod: HttpMethod.post,
      uri: 'admin/users/$userId/suspend',
      queryParams: {'suspend': '$suspend'},
    );
  }

  Future<AdminStats> getStats() {
    return ApiService().request(
      uri: 'admin/stats',
      parser: (res) => AdminStats.fromJson(res as Map<String, dynamic>),
    );
  }
}
