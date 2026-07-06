import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../api/models/user_model.dart';
import '../notifiers/admin_notifier.dart';
import '../notifiers/session_notifier.dart';
import '../widgets/loading_indicator.dart';

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

    if (session.user?.role != Role.admin) {
      return const _UnauthorizedView();
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Admin Panel'),
        elevation: 0,
        bottom: TabBar(
          controller: _tabs,
          tabs: const [
            Tab(icon: Icon(Icons.people), text: 'Users'),
            Tab(icon: Icon(Icons.bar_chart), text: 'Stats'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabs,
        children: const [_UsersTab(), _StatsTab()],
      ),
    );
  }
}

// ── Users tab ───────────────────────────────────────────────────────────────

class _UsersTab extends StatelessWidget {
  const _UsersTab();

  @override
  Widget build(BuildContext context) {
    final admin = context.watch<AdminNotifier>();

    if (admin.usersStatus == AdminStatus.loading) {
      return const LoadingIndicator(message: 'Loading users...');
    }

    if (admin.usersStatus == AdminStatus.error) {
      return _ErrorView(
        message: admin.error,
        onRetry: () => context.read<AdminNotifier>().loadUsers(),
      );
    }

    if (admin.users.isEmpty) {
      return const Center(child: Text('No users found'));
    }

    return RefreshIndicator(
      onRefresh: () => context.read<AdminNotifier>().loadUsers(),
      child: ListView.separated(
        padding: const EdgeInsets.all(16),
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
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: colorScheme.primaryContainer,
          child: Text(
            user.email.substring(0, 1).toUpperCase(),
            style: TextStyle(color: colorScheme.onPrimaryContainer),
          ),
        ),
        title: Text(user.email),
        subtitle: Text(
          user.isSuspended
              ? 'Suspendu'
              : (user.fullName.trim().isEmpty ? '-' : user.fullName),
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            _RoleBadge(role: user.role),
            const SizedBox(width: 8),
            PopupMenuButton<_UserAction>(
              onSelected: (action) => _handleAction(context, action, user),
              itemBuilder: (_) => [
                const PopupMenuItem(
                  value: _UserAction.changeRole,
                  child: ListTile(
                    leading: Icon(Icons.manage_accounts),
                    title: Text('Change role'),
                    contentPadding: EdgeInsets.zero,
                  ),
                ),
                const PopupMenuItem(
                  value: _UserAction.suspend,
                  child: ListTile(
                    leading: Icon(Icons.block),
                    title: Text('Suspend'),
                    contentPadding: EdgeInsets.zero,
                  ),
                ),
                const PopupMenuItem(
                  value: _UserAction.reactivate,
                  child: ListTile(
                    leading: Icon(Icons.check_circle_outline),
                    title: Text('Reactivate'),
                    contentPadding: EdgeInsets.zero,
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
          title: Text('Change role — ${user.email}'),
          content: RadioGroup<Role>(
            groupValue: selected,
            onChanged: (v) => setDialogState(() => selected = v!),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [Role.user, Role.diffuseur, Role.admin].map((r) {
                return RadioListTile<Role>(title: Text(r.apiValue), value: r);
              }).toList(),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () {
                admin.updateRole(user.id, selected);
                Navigator.pop(ctx);
              },
              child: const Text('Confirm'),
            ),
          ],
        ),
      ),
    );
  }
}

enum _UserAction { changeRole, suspend, reactivate }

class _RoleBadge extends StatelessWidget {
  final Role role;
  const _RoleBadge({required this.role});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final (bg, fg) = switch (role) {
      Role.admin => (colorScheme.error, colorScheme.onError),
      Role.diffuseur => (colorScheme.primary, colorScheme.onPrimary),
      _ => (colorScheme.surfaceContainerHighest, colorScheme.onSurfaceVariant),
    };

    return Chip(
      label: Text(role.apiValue, style: TextStyle(color: fg, fontSize: 11)),
      backgroundColor: bg,
      padding: EdgeInsets.zero,
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
    );
  }
}

// ── Stats tab ────────────────────────────────────────────────────────────────

class _StatsTab extends StatelessWidget {
  const _StatsTab();

  @override
  Widget build(BuildContext context) {
    final admin = context.watch<AdminNotifier>();

    if (admin.statsStatus == AdminStatus.loading) {
      return const LoadingIndicator(message: 'Loading stats...');
    }

    if (admin.statsStatus == AdminStatus.error) {
      return _ErrorView(
        message: admin.error,
        onRetry: () => context.read<AdminNotifier>().loadStats(),
      );
    }

    final stats = admin.stats;
    if (stats == null) {
      return const Center(child: Text('No data'));
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _StatCard(
            label: 'Total users',
            value: '${stats.totalUsers}',
            icon: Icons.group,
          ),
          const SizedBox(height: 12),
          ...stats.byRole.entries.map(
            (e) => Padding(
              padding: const EdgeInsets.only(bottom: 12),
              child: _StatCard(
                label: e.key,
                value: '${e.value}',
                icon: _iconForRole(e.key),
              ),
            ),
          ),
        ],
      ),
    );
  }

  IconData _iconForRole(String role) => switch (role) {
    'admin' => Icons.admin_panel_settings,
    'diffuseur' => Icons.mic,
    _ => Icons.person,
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
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Icon(icon, size: 32, color: colorScheme.primary),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(label, style: Theme.of(context).textTheme.bodySmall),
                  Text(
                    value,
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
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

// ── Shared ───────────────────────────────────────────────────────────────────

class _ErrorView extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorView({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.error_outline,
              size: 48,
              color: Theme.of(context).colorScheme.error,
            ),
            const SizedBox(height: 16),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
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
    return Scaffold(
      appBar: AppBar(title: const Text('Admin Panel')),
      body: const Center(child: Text('Access restricted to admins.')),
    );
  }
}
