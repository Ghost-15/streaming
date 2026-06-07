enum UserRole {
  anon('ROLE_ANON'),
  user('ROLE_USER'),
  diffuseur('ROLE_DIFFUSEUR'),
  admin('ROLE_ADMIN');

  final String value;
  const UserRole(this.value);

  static UserRole fromValue(String value) {
    switch (value) {
      case 'ROLE_ADMIN':
        return UserRole.admin;
      case 'ROLE_DIFFUSEUR':
        return UserRole.diffuseur;
      case 'ROLE_USER':
        return UserRole.user;
      default:
        return UserRole.anon;
    }
  }
}

class UserModel {
  final String id;
  final String email;
  final String firstName;
  final String lastName;
  final UserRole role;

  const UserModel({
    required this.id,
    required this.email,
    required this.firstName,
    required this.lastName,
    required this.role,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'],
      email: json['email'],
      firstName: json['firstName'] ?? json['first_name'] ?? '',
      lastName: json['lastName'] ?? json['last_name'] ?? '',
      role: UserRole.fromValue(json['role'] ?? 'ROLE_USER'),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'email': email,
      'firstName': firstName,
      'lastName': lastName,
      'role': role.value,
    };
  }

  String get fullName => '$firstName $lastName';
  bool get isAdmin => role == UserRole.admin;
  bool get isDiffuseur => role == UserRole.diffuseur || role == UserRole.admin;
}
