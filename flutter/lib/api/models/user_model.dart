import 'role.dart';

class UserModel {
  final String id;
  final String email;
  final String firstName;
  final String lastName;
  final Role role;

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
      role: Role.fromValue(json['role'] ?? 'ROLE_USER'),
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
  bool get isAdmin => role == Role.admin;
  bool get isDiffuseur => role == Role.diffuseur || role == Role.admin;
}
