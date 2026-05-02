import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/user.dart';

enum AuthStatus { idle, loading, error }

// ── Current user ──────────────────────────────────────────────────────────────

class CurrentUserNotifier extends Notifier<User?> {
  @override
  User? build() => null;
}

final currentUserProvider =
    NotifierProvider<CurrentUserNotifier, User?>(CurrentUserNotifier.new);

// ── JWT token ─────────────────────────────────────────────────────────────────

class AuthTokenNotifier extends Notifier<String?> {
  @override
  String? build() => null;
}

final authTokenProvider =
    NotifierProvider<AuthTokenNotifier, String?>(AuthTokenNotifier.new);

// ── Auth status ───────────────────────────────────────────────────────────────

class AuthStatusNotifier extends Notifier<AuthStatus> {
  @override
  AuthStatus build() => AuthStatus.idle;
}

final authStatusProvider =
    NotifierProvider<AuthStatusNotifier, AuthStatus>(AuthStatusNotifier.new);
