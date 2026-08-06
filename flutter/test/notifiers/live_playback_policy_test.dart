import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/notifiers/live_playback_policy.dart';

void main() {
  group('liveSeekTarget', () {
    test('never rewinds when playback is just past the buffered edge', () {
      final target = liveSeekTarget(
        currentTime: 75.55,
        rangeStart: 30.5,
        rangeEnd: 75.5,
        maxLatency: 2.5,
        edgeOffset: 0.75,
      );

      expect(target, isNull);
    });

    test('does not seek while playback latency is acceptable', () {
      final target = liveSeekTarget(
        currentTime: 73.5,
        rangeStart: 30.5,
        rangeEnd: 75.5,
        maxLatency: 2.5,
        edgeOffset: 0.75,
      );

      expect(target, isNull);
    });

    test('catches up forwards when playback falls behind', () {
      final target = liveSeekTarget(
        currentTime: 70,
        rangeStart: 30.5,
        rangeEnd: 75.5,
        maxLatency: 2.5,
        edgeOffset: 0.75,
      );

      expect(target, closeTo(74.75, 0.001));
      expect(target!, greaterThan(70));
    });

    test('recovers forwards from an evicted buffered range', () {
      final target = liveSeekTarget(
        currentTime: 25,
        rangeStart: 30.5,
        rangeEnd: 75.5,
        maxLatency: 2.5,
        edgeOffset: 0.75,
      );

      expect(target, closeTo(74.75, 0.001));
      expect(target!, greaterThan(25));
    });

    test('keeps the target inside a very short buffered range', () {
      final target = liveSeekTarget(
        currentTime: 9,
        rangeStart: 10,
        rangeEnd: 10.1,
        maxLatency: 2.5,
        edgeOffset: 0.75,
      );

      expect(target, closeTo(10.05, 0.001));
    });
  });
}
