import 'dart:math' as math;

/// Returns a position near the live edge when playback needs to catch up.
///
/// A live player must never seek backwards automatically: the media clock can
/// briefly move past the buffered edge while the next chunk is being appended.
/// Rewinding in that state would replay the last audio fragment on every small
/// buffer underrun.
double? liveSeekTarget({
  required double currentTime,
  required double rangeStart,
  required double rangeEnd,
  required double maxLatency,
  required double edgeOffset,
}) {
  if (!currentTime.isFinite ||
      !rangeStart.isFinite ||
      !rangeEnd.isFinite ||
      rangeEnd <= rangeStart) {
    return null;
  }

  final isBehindBufferedRange = currentTime < rangeStart;
  final latency = rangeEnd - currentTime;
  if (!isBehindBufferedRange && latency <= maxLatency) return null;

  final target = math.min(
    math.max(rangeStart, rangeEnd - 0.05),
    math.max(rangeStart + 0.05, rangeEnd - edgeOffset),
  );

  // Besides preventing audible repetition, this margin avoids tiny seeks that
  // can destabilise playback without producing a useful latency correction.
  const minForwardSeekDeltaSeconds = 0.5;
  return target > currentTime + minForwardSeekDeltaSeconds ? target : null;
}