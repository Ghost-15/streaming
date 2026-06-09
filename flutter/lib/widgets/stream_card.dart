import 'package:flutter/material.dart';

import '../api/models/stream_model.dart';

class StreamCard extends StatelessWidget {
  final StreamModel stream;
  final VoidCallback? onTap;
  final VoidCallback? onPlay;

  const StreamCard({
    super.key,
    required this.stream,
    this.onTap,
    this.onPlay,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: colorScheme.outline.withValues(alpha: 0.2)),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                stream.title,
                style: textTheme.titleMedium,
                semanticsLabel: 'Stream: ${stream.title}',
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 8),
              Text(
                'by ${stream.broadcasterName}',
                style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant),
                semanticsLabel: 'Broadcaster: ${stream.broadcasterName}',
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Semantics(
                    label: '${stream.listenerCount} listeners',
                    child: Row(
                      children: [
                        Icon(Icons.person, size: 16, color: colorScheme.onSurfaceVariant),
                        const SizedBox(width: 4),
                        Text('${stream.listenerCount}', style: textTheme.bodySmall),
                      ],
                    ),
                  ),
                  Semantics(
                    button: true,
                    enabled: true,
                    label: 'Play stream ${stream.title}',
                    child: SizedBox(
                      width: 48,
                      height: 48,
                      child: Material(
                        color: colorScheme.primary,
                        borderRadius: BorderRadius.circular(24),
                        child: InkWell(
                          onTap: onPlay,
                          borderRadius: BorderRadius.circular(24),
                          child: Icon(Icons.play_arrow, color: colorScheme.onPrimary),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              if (stream.isLive)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Semantics(
                    label: 'Live stream',
                    child: Chip(
                      label: const Text('LIVE'),
                      backgroundColor: colorScheme.error,
                      labelStyle: TextStyle(color: colorScheme.onError),
                      padding: const EdgeInsets.symmetric(horizontal: 8),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
