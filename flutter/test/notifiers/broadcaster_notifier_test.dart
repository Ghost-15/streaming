import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/notifiers/broadcaster_notifier.dart';
import 'package:streampulse/api/repositories/stream_repository.dart';

void main() {
  group('BroadcasterNotifier', () {
    late BroadcasterNotifier notifier;

    setUp(() {
      notifier = BroadcasterNotifier(const StreamRepository());
    });

    tearDown(() {
      notifier.dispose();
    });

    test('initial state is idle', () {
      expect(notifier.state, BroadcasterState.idle);
      expect(notifier.isStreaming, isFalse);
      expect(notifier.isLoading, isFalse);
      expect(notifier.hasError, isFalse);
    });

    test('initial currentStream is null', () {
      expect(notifier.currentStream, isNull);
    });

    test('startStream with empty title sets error state', () async {
      await notifier.startStream('');
      expect(notifier.state, BroadcasterState.error);
      expect(notifier.errorMessage, isNotEmpty);
    });

    test('clearError resets to idle', () async {
      await notifier.startStream('');
      expect(notifier.hasError, isTrue);
      notifier.clearError();
      expect(notifier.state, BroadcasterState.idle);
      expect(notifier.errorMessage, isEmpty);
    });

    test('stopStream does nothing when no current stream', () async {
      await notifier.stopStream();
      expect(notifier.state, BroadcasterState.idle);
    });

    test('notifies listeners on clearError', () {
      bool notified = false;
      notifier.addListener(() => notified = true);
      notifier.clearError();
      expect(notified, isTrue);
    });
  });
}
