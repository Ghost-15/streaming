import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/notifiers/audio_notifier.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('AudioNotifier', () {
    late AudioNotifier notifier;

    setUp(() {
      notifier = AudioNotifier();
    });

    tearDown(() {
      notifier.dispose();
    });

    test('initial state is idle', () {
      expect(notifier.playbackState, AudioPlaybackState.idle);
      expect(notifier.isPlaying, isFalse);
      expect(notifier.isPaused, isFalse);
      expect(notifier.isLoading, isFalse);
      expect(notifier.hasError, isFalse);
    });

    test('initial volume is 1.0', () {
      expect(notifier.volume, 1.0);
    });

    test('initial position and duration are zero', () {
      expect(notifier.position, Duration.zero);
      expect(notifier.duration, Duration.zero);
    });

    test('no stream selected initially', () {
      expect(notifier.currentStream, isNull);
    });

    test('toggleShuffle changes isShuffled', () {
      expect(notifier.isShuffled, isFalse);
      notifier.toggleShuffle();
      expect(notifier.isShuffled, isTrue);
      notifier.toggleShuffle();
      expect(notifier.isShuffled, isFalse);
    });

    test('toggleLoop changes isLooping', () {
      expect(notifier.isLooping, isFalse);
      notifier.toggleLoop();
      expect(notifier.isLooping, isTrue);
      notifier.toggleLoop();
      expect(notifier.isLooping, isFalse);
    });

    test('clearError resets error message', () {
      notifier.clearError();
      expect(notifier.errorMessage, isEmpty);
    });

    test('progress is 0 when duration is zero', () {
      expect(notifier.progress, 0.0);
    });

    test('notifies listeners on toggleShuffle', () {
      bool notified = false;
      notifier.addListener(() => notified = true);
      notifier.toggleShuffle();
      expect(notified, isTrue);
    });

    test('notifies listeners on toggleLoop', () {
      bool notified = false;
      notifier.addListener(() => notified = true);
      notifier.toggleLoop();
      expect(notified, isTrue);
    });
  });
}
