import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../api/models/user_model.dart';
import '../notifiers/admin_notifier.dart';
import '../notifiers/session_notifier.dart';
import '../widgets/loading_indicator.dart';
import '../widgets/page_header.dart';

class AdminScreen extends StatefulWidget {
  const AdminScreen({super.key});

  @override
  State<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends State<AdminScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final admin = context.read<AdminNotifier>();
      admin.loadUsers();
      admin.loadStats();
    });
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();
    final cs = Theme.of(context).colorScheme;

    if (session.user?.role != Role.admin) {
      return const _UnauthorizedView();
    }

    return Scaffold(
      body: Column(
        children: [
          PageHeader(
            icon: Icons.admin_panel_settings_rounded,
            title: 'Administration',
            subtitle: 'Gestion des utilisateurs et statistiques',
            actions: [
              IconButton(
                icon: Icon(
                  Icons.logout_rounded,
                  size: 20,
                  color: cs.onSurfaceVariant,
                ),
                tooltip: 'Se déconnecter',
                onPressed: () => session.logout(),
              ),
            ],
          ),
          const SizedBox(height: 8),
          TabBar(
            controller: _tabs,
            tabs: const [
              Tab(icon: Icon(Icons.people_rounded), text: 'Utilisateurs'),
              Tab(icon: Icon(Icons.bar_chart_rounded), text: 'Statistiques'),
            ],
            indicatorPadding: const EdgeInsets.symmetric(horizontal: 16),
            dividerColor: cs.outlineVariant,
          ),
          Expanded(
            child: TabBarView(
              controller: _tabs,
              children: const [_UsersTab(), _StatsTab()],
            ),
          ),
        ],
      ),
    );
  }
}

// ── Users tab ─────────────────────────────────────────────────────────────────

class _UsersTab extends StatelessWidget {
  const _UsersTab();

  @override
  Widget build(BuildContext context) {
    final admin = context.watch<AdminNotifier>();

    if (admin.usersStatus == AdminStatus.loading) {
      return const LoadingIndicator(message: 'Chargement des utilisateurs…');
    }
    if (admin.usersStatus == AdminStatus.error) {
      return _ErrorView(
        message: admin.error,
        onRetry: () => context.read<AdminNotifier>().loadUsers(),
      );
    }
    if (admin.users.isEmpty) {
      return const Center(child: Text('Aucun utilisateur trouvé'));
    }

    return RefreshIndicator(
      onRefresh: () => context.read<AdminNotifier>().loadUsers(),
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
        itemCount: admin.users.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (_, i) => _UserTile(user: admin.users[i]),
      ),
    );
  }
}

class _UserTile extends StatelessWidget {
  final UserModel user;
  const _UserTile({required this.user});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final initial = user.firstName.isNotEmpty
        ? user.firstName[0].toUpperCase()
        : user.email[0].toUpperCase();

    return Container(
      decoration: BoxDecoration(
        color: cs.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: cs.outlineVariant, width: 0.8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.12),
            blurRadius: 10,
            offset: const Offset(0, 3),
            spreadRadius: -3,
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            CircleAvatar(
              radius: 20,
              backgroundColor: cs.primaryContainer,
              child: Text(
                initial,
                style: tt.titleSmall?.copyWith(
                  color: cs.onPrimaryContainer,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    user.fullName.trim().isNotEmpty
                        ? user.fullName
                        : user.email,
                    style: tt.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  Text(
                    user.email,
                    style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            if (user.isSuspended) ...[
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                decoration: BoxDecoration(
                  color: cs.errorContainer,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Text(
                  'Suspendu',
                  style: TextStyle(
                    fontSize: 9,
                    color: cs.onErrorContainer,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
              const SizedBox(width: 6),
            ],
            _RoleBadge(role: user.role),
            PopupMenuButton<_UserAction>(
              icon: Icon(
                Icons.more_vert_rounded,
                size: 18,
                color: cs.onSurfaceVariant,
              ),
              onSelected: (action) => _handleAction(context, action, user),
              itemBuilder: (_) => const [
                PopupMenuItem(
                  value: _UserAction.changeRole,
                  child: ListTile(
                    leading: Icon(Icons.manage_accounts_rounded),
                    title: Text('Changer le rôle'),
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                  ),
                ),
                PopupMenuItem(
                  value: _UserAction.suspend,
                  child: ListTile(
                    leading: Icon(Icons.block_rounded),
                    title: Text('Suspendre'),
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                  ),
                ),
                PopupMenuItem(
                  value: _UserAction.reactivate,
                  child: ListTile(
                    leading: Icon(Icons.check_circle_outline_rounded),
                    title: Text('Réactiver'),
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  void _handleAction(BuildContext context, _UserAction action, UserModel user) {
    switch (action) {
      case _UserAction.changeRole:
        _showRoleDialog(context, user);
      case _UserAction.suspend:
        context.read<AdminNotifier>().suspendUser(user.id, suspend: true);
      case _UserAction.reactivate:
        context.read<AdminNotifier>().suspendUser(user.id, suspend: false);
    }
  }

  void _showRoleDialog(BuildContext context, UserModel user) {
    final admin = context.read<AdminNotifier>();
    Role selected = user.role;

    showDialog<void>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          title: Text('Rôle — ${user.email}'),
          content: RadioGroup<Role>(
            groupValue: selected,
            onChanged: (v) => setDialogState(() => selected = v!),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [Role.user, Role.diffuseur, Role.admin].map((r) {
                return RadioListTile<Role>(
                  title: Text(_roleLabel(r)),
                  value: r,
                );
              }).toList(),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Annuler'),
            ),
            FilledButton(
              onPressed: () {
                admin.updateRole(user.id, selected);
                Navigator.pop(ctx);
              },
              child: const Text('Confirmer'),
            ),
          ],
        ),
      ),
    );
  }

  String _roleLabel(Role role) => switch (role) {
    Role.admin => 'Administrateur',
    Role.diffuseur => 'Diffuseur',
    _ => 'Auditeur',
  };
}

enum _UserAction { changeRole, suspend, reactivate }

class _RoleBadge extends StatelessWidget {
  final Role role;
  const _RoleBadge({required this.role});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final (bg, fg) = switch (role) {
      Role.admin => (cs.error, cs.onError),
      Role.diffuseur => (cs.primary, cs.onPrimary),
      _ => (cs.surfaceContainerHighest, cs.onSurfaceVariant),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        switch (role) {
          Role.admin => 'Admin',
          Role.diffuseur => 'Diffuseur',
          _ => 'Auditeur',
        },
        style: TextStyle(
          fontSize: 10,
          color: fg,
          fontWeight: FontWeight.w700,
          letterSpacing: 0.2,
        ),
      ),
    );
  }
}

// ── Stats tab ─────────────────────────────────────────────────────────────────

class _StatsTab extends StatelessWidget {
  const _StatsTab();

  @override
  Widget build(BuildContext context) {
    final admin = context.watch<AdminNotifier>();

    if (admin.statsStatus == AdminStatus.loading) {
      return const LoadingIndicator(message: 'Chargement des statistiques…');
    }
    if (admin.statsStatus == AdminStatus.error) {
      return _ErrorView(
        message: admin.error,
        onRetry: () => context.read<AdminNotifier>().loadStats(),
      );
    }

    final stats = admin.stats;
    if (stats == null) {
      return const Center(child: Text('Aucune donnée disponible'));
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _StatCard(
            label: 'Utilisateurs au total',
            value: '${stats.totalUsers}',
            icon: Icons.group_rounded,
          ),
          const SizedBox(height: 12),
          ...stats.byRole.entries.map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: _StatCard(
                label: _roleLabel(e.key),
                value: '${e.value}',
                icon: _iconForRole(e.key),
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _roleLabel(String role) => switch (role) {
    'admin' => 'Administrateurs',
    'diffuseur' => 'Diffuseurs',
    _ => 'Auditeurs',
  };

  IconData _iconForRole(String role) => switch (role) {
    'admin' => Icons.admin_panel_settings_rounded,
    'diffuseur' => Icons.mic_rounded,
    _ => Icons.headphones_rounded,
  };
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final IconData icon;

  const _StatCard({
    required this.label,
    required this.value,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Container(
      decoration: BoxDecoration(
        color: cs.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: cs.outlineVariant, width: 0.8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.14),
            blurRadius: 12,
            offset: const Offset(0, 3),
            spreadRadius: -4,
          ),
          BoxShadow(
            color: cs.primary.withValues(alpha: 0.05),
            blurRadius: 16,
            spreadRadius: -2,
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: cs.primary.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: cs.primary.withValues(alpha: 0.15),
                  width: 1,
                ),
              ),
              child: Icon(icon, size: 24, color: cs.primary),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    value,
                    style: tt.headlineMedium?.copyWith(
                      fontWeight: FontWeight.w800,
                      letterSpacing: -0.5,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Shared ────────────────────────────────────────────────────────────────────

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline_rounded, size: 48, color: cs.error),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
            ),
            const SizedBox(height: 20),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh_rounded, size: 18),
              label: const Text('Réessayer'),
            ),
          ],
        ),
      ),
    );
  }
}

class _UnauthorizedView extends StatelessWidget {
  const _UnauthorizedView();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Scaffold(
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(40),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: cs.surfaceContainerHigh,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  Icons.lock_outline_rounded,
                  size: 36,
                  color: cs.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 20),
              Text(
                'Accès restreint',
                style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text(
                'Cette page est réservée aux administrateurs.',
                style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
