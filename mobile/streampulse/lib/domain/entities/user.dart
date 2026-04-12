enum UserRole { anon, user, diffuseur, admin }

class User {
  final String id;
  final String email;
  final String firstName;
  final String lastName;
  final UserRole role;

  const User({
    required this.id,
    required this.email,
    required this.firstName,
    required this.lastName,
    required this.role,
  });

  String get fullName => '$firstName $lastName';
  bool get isAdmin => role == UserRole.admin;
  bool get isDiffuseur => role == UserRole.diffuseur || role == UserRole.admin;
}
