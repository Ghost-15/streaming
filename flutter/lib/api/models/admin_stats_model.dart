class AdminStats {
  final int totalUsers;
  final Map<String, int> byRole;

  const AdminStats({required this.totalUsers, required this.byRole});

  factory AdminStats.fromJson(Map<String, dynamic> json) {
    return AdminStats(
      totalUsers: json['total_users'] as int? ?? 0,
      byRole: (json['by_role'] as Map<String, dynamic>?)
              ?.map((k, v) => MapEntry(k, v as int)) ??
          {},
    );
  }
}
