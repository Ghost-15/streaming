import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../notifiers/broadcaster_notifier.dart';
import '../notifiers/session_notifier.dart';
import '../widgets/loading_indicator.dart';

class BroadcasterScreen extends StatefulWidget {
  const BroadcasterScreen({super.key});

  @override
  State<BroadcasterScreen> createState() => _BroadcasterScreenState();
}

class _BroadcasterScreenState extends State<BroadcasterScreen> {
  final _titleController = TextEditingController();

  @override
  void dispose() {
    _titleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();
    final broadcaster = context.watch<BroadcasterNotifier>();

    final user = session.user;
    if (user == null ||
        (user.role != Role.diffuseur && user.role != Role.admin)) {
      return const _UnauthorizedView();
    }

    return Scaffold(
      appBar: AppBar(title: const Text('Broadcaster'), elevation: 0),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _StatusCard(
              state: broadcaster.state,
              listenerCount: broadcaster.listenerCount,
              streamTitle: broadcaster.currentStream?.title,
            ),
            const SizedBox(height: 24),
            if (!broadcaster.isStreaming) ...[
              TextField(
                controller: _titleController,
                enabled: !broadcaster.isLoading,
                decoration: const InputDecoration(
                  labelText: 'Stream title',
                  hintText: 'Enter a title for your live stream',
                  border: OutlineInputBorder(),
                  prefixIcon: Icon(Icons.mic),
                ),
                maxLength: 100,
              ),
            ],
            if (broadcaster.hasError) ...[
              const SizedBox(height: 8),
              _ErrorBanner(
                message: broadcaster.errorMessage,
                onDismiss: () => context.read<BroadcasterNotifier>().clearError(),
              ),
            ],
            const Spacer(),
            if (broadcaster.isLoading)
              const LoadingIndicator(message: 'Please wait...')
            else
              _ToggleButton(
                isStreaming: broadcaster.isStreaming,
                onStart: () => context
                    .read<BroadcasterNotifier>()
                    .startStream(_titleController.text.trim()),
                onStop: () =>
                    context.read<BroadcasterNotifier>().stopStream(),
              ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }
}

class _StatusCard extends StatelessWidget {
  final BroadcasterState state;
  final int listenerCount;
  final String? streamTitle;

  const _StatusCard({
    required this.state,
    required this.listenerCount,
    this.streamTitle,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isLive = state == BroadcasterState.streaming;

    return Card(
      color: isLive ? colorScheme.errorContainer : colorScheme.surfaceContainerHighest,
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  isLive ? Icons.fiber_manual_record : Icons.radio_button_unchecked,
                  color: isLive ? colorScheme.error : colorScheme.outline,
                  size: 16,
                ),
                const SizedBox(width: 8),
                Text(
                  isLive ? 'LIVE' : 'OFFLINE',
                  style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: isLive ? colorScheme.onErrorContainer : colorScheme.onSurfaceVariant,
                      ),
                ),
              ],
            ),
            if (streamTitle != null) ...[
              const SizedBox(height: 8),
              Text(
                streamTitle!,
                style: Theme.of(context).textTheme.bodyMedium,
                textAlign: TextAlign.center,
              ),
            ],
            if (isLive) ...[
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.people, size: 18),
                  const SizedBox(width: 6),
                  Text(
                    '$listenerCount listener${listenerCount != 1 ? 's' : ''}',
                    style: Theme.of(context).textTheme.bodyLarge,
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _ToggleButton extends StatelessWidget {
  final bool isStreaming;
  final VoidCallback onStart;
  final VoidCallback onStop;

  const _ToggleButton({
    required this.isStreaming,
    required this.onStart,
    required this.onStop,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    if (isStreaming) {
      return FilledButton.icon(
        onPressed: onStop,
        icon: const Icon(Icons.stop),
        label: const Text('Stop Stream'),
        style: FilledButton.styleFrom(
          backgroundColor: colorScheme.error,
          foregroundColor: colorScheme.onError,
          minimumSize: const Size.fromHeight(56),
        ),
      );
    }

    return FilledButton.icon(
      onPressed: onStart,
      icon: const Icon(Icons.fiber_manual_record),
      label: const Text('Go Live'),
      style: FilledButton.styleFrom(
        minimumSize: const Size.fromHeight(56),
      ),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  final String message;
  final VoidCallback onDismiss;

  const _ErrorBanner({required this.message, required this.onDismiss});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      color: colorScheme.errorContainer,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        child: Row(
          children: [
            Icon(Icons.error_outline, color: colorScheme.error),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                message,
                style: TextStyle(color: colorScheme.onErrorContainer),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.close),
              onPressed: onDismiss,
              color: colorScheme.onErrorContainer,
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
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Broadcaster'), elevation: 0),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.lock_outline, size: 64, color: colorScheme.outline),
            const SizedBox(height: 16),
            Text(
              'Access Denied',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              'You need a broadcaster or admin role to access this page.',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}
