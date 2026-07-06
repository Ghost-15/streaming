enum Role {
  anon('ROLE_ANON'),
  user('ROLE_USER'),
  diffuseur('ROLE_DIFFUSEUR'),
  admin('ROLE_ADMIN');

  final String value;
  const Role(this.value);

  // Value sent to the Go API (lowercase, no prefix).
  String get apiValue => name;

  static Role fromValue(String value) {
    switch (value) {
      case 'ROLE_ADMIN':
      case 'admin':
        return Role.admin;
      case 'ROLE_DIFFUSEUR':
      case 'diffuseur':
        return Role.diffuseur;
      case 'ROLE_USER':
      case 'user':
        return Role.user;
      default:
        return Role.anon;
    }
  }
}
