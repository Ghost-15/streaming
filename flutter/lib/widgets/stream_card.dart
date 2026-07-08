import 'package:flutter/material.dart';

import '../api/models/stream_model.dart';

// Compact = square card for horizontal carousels
// Default = full-width row for lists
class StreamCard extends StatelessWidget {
  final StreamModel stream;
  final bool compact;
  final VoidCallback? onTap;
  final VoidCallback? onPlay;

  const StreamCard({
    super.key,
    required this.stream,
    this.compact = false,
    this.onTap,
    this.onPlay,
  });

  @override
  Widget build(BuildContext context) {
    return compact
        ? _CompactCard(stream: stream, onPlay: onPlay ?? onTap)
        : _RowCard(stream: stream, onTap: onTap, onPlay: onPlay);
  }
}

// ── Compact square card (carousels) ───────────────────────────────────────────

class _CompactCard extends StatelessWidget {
  final StreamModel stream;
  final VoidCallback? onPlay;
  const _CompactCard({required this.stream, this.onPlay});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final accent = _streamAccent(stream.id);

    return GestureDetector(
      onTap: onPlay,
      child: SizedBox(
        width: 148,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(14),
              child: SizedBox(
                width: 148,
                height: 148,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    Container(
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                          colors: [accent, accent.withValues(alpha: 0.5)],
                        ),
                      ),
                    ),
                    // Background icon decoration
                    Positioned(
                      right: -10,
                      bottom: -10,
                      child: Icon(
                        Icons.radio_rounded,
                        size: 90,
                        color: Colors.white.withValues(alpha: 0.1),
                      ),
                    ),
                    // LIVE badge
                    if (stream.isLive)
                      const Positioned(
                        top: 10,
                        left: 10,
                        child: _LiveBadge(),
                      ),
                    // Play button center
                    Center(
                      child: Container(
                        width: 44,
                        height: 44,
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.9),
                          borderRadius: BorderRadius.circular(22),
                        ),
                        child: Icon(Icons.play_arrow_rounded, color: accent, size: 26),
                      ),
                    ),
                    // Listener count
                    Positioned(
                      bottom: 8,
                      right: 10,
                      child: _ListenerCount(count: stream.listenerCount, light: true),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 8),
            Text(
              stream.title,
              style: tt.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            Text(
              stream.broadcasterName,
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }
}

// ── Full-width row card (lists) ───────────────────────────────────────────────

class _RowCard extends StatelessWidget {
  final StreamModel stream;
  final VoidCallback? onTap;
  final VoidCallback? onPlay;
  const _RowCard({required this.stream, this.onTap, this.onPlay});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final accent = _streamAccent(stream.id);

    return Material(
      color: cs.surfaceContainer,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              // Artwork thumbnail
              ClipRRect(
                borderRadius: BorderRadius.circular(10),
                child: SizedBox(
                  width: 56,
                  height: 56,
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      Container(
                        decoration: BoxDecoration(
                          gradient: LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: [accent, accent.withValues(alpha: 0.5)],
                          ),
                        ),
                      ),
                      const Center(
                        child: Icon(Icons.radio_rounded, color: Colors.white54, size: 28),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 14),
              // Info
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      stream.title,
                      style: tt.titleSmall?.copyWith(fontWeight: FontWeight.w600),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    Text(
                      stream.broadcasterName,
                      style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        if (stream.isLive) ...[
                          const _LiveBadge(),
                          const SizedBox(width: 8),
                        ],
                        _ListenerCount(count: stream.listenerCount),
                      ],
                    ),
                  ],
                ),
              ),
              // Play button
              IconButton(
                onPressed: onPlay,
                style: IconButton.styleFrom(
                  backgroundColor: cs.primary,
                  foregroundColor: cs.onPrimary,
                ),
                icon: const Icon(Icons.play_arrow_rounded, size: 22),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Shared ────────────────────────────────────────────────────────────────────

class _LiveBadge extends StatelessWidget {
  const _LiveBadge();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: cs.error,
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        'LIVE',
        style: TextStyle(
          color: cs.onError,
          fontSize: 9,
          fontWeight: FontWeight.w800,
          letterSpacing: 0.5,
        ),
      ),
    );
  }
}

class _ListenerCount extends StatelessWidget {
  final int count;
  final bool light;
  const _ListenerCount({required this.count, this.light = false});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final color = light ? Colors.white70 : cs.onSurfaceVariant;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.headphones_rounded, size: 12, color: color),
        const SizedBox(width: 3),
        Text(
          '$count',
          style: TextStyle(
            fontSize: 11,
            color: color,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }
}

Color _streamAccent(String id) {
  const palette = [
    Color(0xFF6366F1),
    Color(0xFF8B5CF6),
    Color(0xFFEC4899),
    Color(0xFF14B8A6),
    Color(0xFFF59E0B),
    Color(0xFF10B981),
    Color(0xFFEF4444),
    Color(0xFF3B82F6),
    Color(0xFF0EA5E9),
    Color(0xFFF97316),
  ];
  final hash = id.codeUnits.fold(0, (a, b) => a + b);
  return palette[hash % palette.length];
}
