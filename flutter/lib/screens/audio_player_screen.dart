import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../notifiers/audio_notifier.dart';
import '../notifiers/stream_notifier.dart';
import '../widgets/page_header.dart';
import '../widgets/stream_card.dart';

class AudioPlayerScreen extends StatefulWidget {
  const AudioPlayerScreen({super.key});

  @override
  State<AudioPlayerScreen> createState() => _AudioPlayerScreenState();
}

class _AudioPlayerScreenState extends State<AudioPlayerScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<StreamNotifier>().loadActive();
    });
  }

  @override
  Widget build(BuildContext context) {
    final streams = context.watch<StreamNotifier>();
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => context.read<StreamNotifier>().loadActive(),
        child: CustomScrollView(
          slivers: [
            SliverToBoxAdapter(
              child: PageHeader(
                icon: Icons.radio_rounded,
                title: 'Live',
                subtitle: 'Streams en direct',
                actions: [
                  IconButton(
                    icon: const Icon(Icons.refresh_rounded, size: 20),
                    color: cs.onSurfaceVariant,
                    onPressed: () =>
                        context.read<StreamNotifier>().loadActive(),
                  ),
                ],
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 24, 20, 12),
                child: Text(
                  'Streams disponibles',
                  style: tt.titleMedium?.copyWith(fontWeight: FontWeight.w700),
                ),
              ),
            ),
            if (streams.isLoading)
              const SliverFillRemaining(
                child: Center(child: CircularProgressIndicator()),
              )
            else if (streams.error != null && streams.streams.isEmpty)
              SliverFillRemaining(
                child: _ErrorState(
                  message: streams.error!,
                  onRetry: () => context.read<StreamNotifier>().loadActive(),
                ),
              )
            else if (streams.streams.isEmpty)
              const SliverFillRemaining(child: _EmptyState())
            else
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 120),
                sliver: SliverList.separated(
                  itemCount: streams.streams.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 12),
                  itemBuilder: (_, i) => StreamCard(
                    stream: streams.streams[i],
                    onPlay: () => context.read<AudioNotifier>().playStream(
                      streams.streams[i],
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Center(
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
                Icons.radio_outlined,
                size: 36,
                color: cs.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 20),
            Text(
              'Aucun stream en direct',
              style: tt.titleMedium?.copyWith(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 8),
            Text(
              'Reviens plus tard ou invite un diffuseur à lancer un live.',
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorState({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(40),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.wifi_off_rounded, size: 48, color: cs.onSurfaceVariant),
            const SizedBox(height: 16),
            Text(
              'Impossible de charger les streams',
              style: tt.titleSmall?.copyWith(fontWeight: FontWeight.w600),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
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
