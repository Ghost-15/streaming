import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/widgets/audio_controls.dart';

void main() {
  Widget controls({
    required bool isPlaying,
    required bool isLoading,
    required VoidCallback onPlay,
    required VoidCallback onPause,
    required VoidCallback onStop,
  }) {
    return MaterialApp(
      home: Scaffold(
        body: AudioControls(
          isPlaying: isPlaying,
          isLoading: isLoading,
          onPlay: onPlay,
          onPause: onPause,
          onStop: onStop,
        ),
      ),
    );
  }

  testWidgets('play and pause actions remain reversible', (tester) async {
    var playCount = 0;
    var pauseCount = 0;

    await tester.pumpWidget(
      controls(
        isPlaying: true,
        isLoading: false,
        onPlay: () => playCount++,
        onPause: () => pauseCount++,
        onStop: () {},
      ),
    );
    expect(find.byIcon(Icons.pause), findsOneWidget);
    await tester.tap(find.byIcon(Icons.pause));
    expect(pauseCount, 1);

    await tester.pumpWidget(
      controls(
        isPlaying: false,
        isLoading: false,
        onPlay: () => playCount++,
        onPause: () => pauseCount++,
        onStop: () {},
      ),
    );
    expect(find.byIcon(Icons.play_arrow), findsOneWidget);
    await tester.tap(find.byIcon(Icons.play_arrow));
    expect(playCount, 1);
  });

  testWidgets('stop remains available while audio is loading', (tester) async {
    var stopCount = 0;
    await tester.pumpWidget(
      controls(
        isPlaying: false,
        isLoading: true,
        onPlay: () {},
        onPause: () {},
        onStop: () => stopCount++,
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.tap(find.byIcon(Icons.stop));
    expect(stopCount, 1);
  });
}
