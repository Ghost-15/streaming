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

    test('no stream selected initially', () {
      expect(notifier.currentStream, isNull);
    });

    test('clearError resets error message', () {
      notifier.clearError();
      expect(notifier.errorMessage, isEmpty);
    });
  });
}
