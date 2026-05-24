import '../models/user_model.dart';
import '../../domain/entities/user.dart';

/// Mapper to convert UserModel (DTO) to User (domain entity)
class UserMapper {
  static User toDomain(UserModel model) {
    // Map displayName to firstName/lastName or split username
    final names = model.displayName.isEmpty
        ? model.username.split(' ')
        : model.displayName.split(' ');

    final firstName = names.isNotEmpty ? names[0] : model.username;
    final lastName = names.length > 1 ? names.sublist(1).join(' ') : '';

    return User(
      id: model.id,
      email: model.email,
      firstName: firstName,
      lastName: lastName,
      role: _roleFromBroadcaster(model.isBroadcaster),
    );
  }

  static List<User> toDomainList(List<UserModel> models) {
    return models.map(toDomain).toList();
  }

  /// Determine user role based on isBroadcaster flag
  static UserRole _roleFromBroadcaster(bool isBroadcaster) {
    return isBroadcaster ? UserRole.diffuseur : UserRole.user;
  }
}
