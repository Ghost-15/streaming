// Mobile — list of active streams + join button.
// Sprint 1 — US-003, Sprint 2 — US-009.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/stream_provider.dart';

class StreamPage extends ConsumerWidget {
  const StreamPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final streams = ref.watch(activeStreamsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Live Streams')),
      body: streams.isEmpty
          ? const Center(child: Text('No active streams'))
          : ListView.builder(
              itemCount: streams.length,
              itemBuilder: (_, i) => ListTile(
                title: Text(streams[i].title),
                subtitle: Text('${streams[i].listenerCount} listeners'),
                trailing: const Icon(Icons.play_circle_outline),
                onTap: () {
                  // TODO Sprint 2 — US-009: navigate to player
                },
              ),
            ),
    );
  }
}
