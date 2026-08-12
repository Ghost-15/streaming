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

    return Semantics(
      button: true,
      label: 'Écouter ${stream.title} par ${stream.broadcasterName}',
      child: GestureDetector(
        onTap: onPlay,
        child: SizedBox(
          width: 148,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Thumbnail with glass border
            Container(
              width: 148,
              height: 148,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: accent.withValues(alpha: 0.25),
                  width: 1,
                ),
                boxShadow: [
                  BoxShadow(
                    color: accent.withValues(alpha: 0.18),
                    blurRadius: 16,
                    offset: const Offset(0, 4),
                    spreadRadius: -4,
                  ),
                ],
              ),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(15),
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    // Gradient fill
                    Container(
                      decoration: BoxDecoration(
                        gradient: LinearGradient(
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                          colors: [
                            accent,
                            accent.withValues(alpha: 0.45),
                            cs.surfaceContainerHigh,
                          ],
                          stops: const [0.0, 0.55, 1.0],
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
                        color: Colors.white.withValues(alpha: 0.08),
                      ),
                    ),
                    // LIVE badge
                    if (stream.isLive)
                      const Positioned(top: 10, left: 10, child: _LiveBadge()),
                    // Play button
                    Center(
                      child: Container(
                        width: 44,
                        height: 44,
                        decoration: BoxDecoration(
                          color: Colors.white.withValues(alpha: 0.88),
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withValues(alpha: 0.2),
                              blurRadius: 8,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: Icon(
                          Icons.play_arrow_rounded,
                          color: accent,
                          size: 26,
                        ),
                      ),
                    ),
                    // Listener count
                    Positioned(
                      bottom: 8,
                      right: 10,
                      child: _ListenerCount(
                        count: stream.listenerCount,
                        light: true,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 9),
            Text(
              stream.title,
              style: tt.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
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
          ],
        ),
      ),
    ),
    );
  }
}

// ── Full-width row card (lists) — glassmorphism ───────────────────────────────

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

    return Semantics(
      label: '${stream.title} par ${stream.broadcasterName}'
          '${stream.isLive ? ", en direct" : ""}'
          ', ${stream.listenerCount} auditeur${stream.listenerCount != 1 ? "s" : ""}',
      child: GestureDetector(
        onTap: onTap ?? onPlay,
        child: Container(
        decoration: BoxDecoration(
          color: cs.surfaceContainerHigh,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: cs.outlineVariant, width: 0.8),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.18),
              blurRadius: 14,
              offset: const Offset(0, 4),
              spreadRadius: -4,
            ),
            // Accent glow tinted to stream color — the "crystal" effect
            BoxShadow(
              color: accent.withValues(alpha: 0.07),
              blurRadius: 20,
              spreadRadius: -2,
            ),
          ],
        ),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              // Thumbnail
              Container(
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(
                      color: accent.withValues(alpha: 0.3),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
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
                          child: Icon(
                            Icons.radio_rounded,
                            color: Colors.white54,
                            size: 26,
                          ),
                        ),
                      ],
                    ),
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
                      style: tt.titleSmall?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
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

              // Play button — circle with subtle shadow
              const SizedBox(width: 8),
              Semantics(
                button: true,
                label: 'Écouter ${stream.title}',
                excludeSemantics: true,
                child: GestureDetector(
                  onTap: onPlay,
                  child: Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: cs.primary,
                      shape: BoxShape.circle,
                      boxShadow: [
                        BoxShadow(
                          color: cs.primary.withValues(alpha: 0.35),
                          blurRadius: 10,
                          offset: const Offset(0, 3),
                        ),
                      ],
                    ),
                    child: Icon(
                      Icons.play_arrow_rounded,
                      color: cs.onPrimary,
                      size: 22,
                    ),
                  ),
                ),
              ),
            ],
          ),
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
    return Semantics(
      label: 'En direct',
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
        decoration: BoxDecoration(
          color: cs.error,
          borderRadius: BorderRadius.circular(6),
          boxShadow: [
            BoxShadow(
              color: cs.error.withValues(alpha: 0.35),
              blurRadius: 6,
              offset: const Offset(0, 1),
            ),
          ],
        ),
        child: Text(
          'LIVE',
          style: TextStyle(
            color: cs.onError,
            fontSize: 9,
            fontWeight: FontWeight.w800,
            letterSpacing: 0.8,
          ),
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

    return Semantics(
      label: '$count auditeur${count != 1 ? "s" : ""}',
      child: Row(
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
      ),
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
