enum Role {
  anon('ROLE_ANON'),
  user('ROLE_USER'),
  diffuseur('ROLE_DIFFUSEUR'),
  admin('ROLE_ADMIN');

  final String value;
  const Role(this.value);

  static Role fromValue(String value) {
    switch (value) {
      case 'ROLE_ADMIN':
        return Role.admin;
      case 'ROLE_DIFFUSEUR':
        return Role.diffuseur;
      case 'ROLE_USER':
        return Role.user;
      default:
        return Role.anon;
    }
  }
}
