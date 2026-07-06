import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../notifiers/session_notifier.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('StreamPulse'),
        centerTitle: true,
        actions: [
          if (session.isAuthenticated)
            IconButton(
              icon: const Icon(Icons.logout),
              tooltip: 'Déconnexion',
              onPressed: () => session.logout(),
            )
          else
            TextButton(
              onPressed: () => context.push('/login'),
              child: const Text('Se connecter'),
            ),
        ],
      ),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                'StreamPulse',
                style: Theme.of(context).textTheme.displaySmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                'Live Audio Streaming',
                style: Theme.of(context).textTheme.bodyLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 40),
              FilledButton.icon(
                onPressed: () => context.push('/player'),
                icon: const Icon(Icons.play_circle),
                label: const Text('Écouter un stream'),
                style: FilledButton.styleFrom(
                  minimumSize: const Size.fromHeight(52),
                ),
              ),
              if (session.isAuthenticated) ...[
                const SizedBox(height: 16),
                _DashboardButton(role: session.user!.role),
              ],
              if (!session.isAuthenticated) ...[
                const SizedBox(height: 24),
                Text(
                  'Connectez-vous pour diffuser ou accéder au panel admin.',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                  textAlign: TextAlign.center,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _DashboardButton extends StatelessWidget {
  final Role role;
  const _DashboardButton({required this.role});

  @override
  Widget build(BuildContext context) {
    return switch (role) {
      Role.admin => FilledButton.tonalIcon(
          onPressed: () => context.push('/admin'),
          icon: const Icon(Icons.admin_panel_settings),
          label: const Text('Panel admin'),
          style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
        ),
      Role.diffuseur => FilledButton.tonalIcon(
          onPressed: () => context.push('/broadcaster'),
          icon: const Icon(Icons.mic),
          label: const Text('Espace diffuseur'),
          style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
        ),
      _ => FilledButton.tonalIcon(
          onPressed: () => context.push('/profile'),
          icon: const Icon(Icons.person),
          label: const Text('Mon profil'),
          style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
        ),
    };
  }
}
