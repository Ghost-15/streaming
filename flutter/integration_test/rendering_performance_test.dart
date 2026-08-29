import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

import 'package:streampulse/api/models/stream_model.dart';
import 'package:streampulse/widgets/stream_card.dart';

// Rendering budget required by the project brief: the interface must stay fluid
// (60 FPS) while real-time data keeps arriving. At 60 FPS the platform hands the
// framework one frame every 16.67 ms, split between the UI thread (build and
// layout) and the raster thread (painting).
const double frameBudgetMs = 1000 / 60;

// A live listener receives audio chunks roughly twice a second and a listener
// count update on top. This ticker is deliberately faster than production so the
// UI is measured under more pressure than it will actually face.
const Duration dataTickInterval = Duration(milliseconds: 16);

/// Records the real frame timings reported by the engine.
class _FrameRecorder {
  final List<double> uiMillis = [];
  final List<double> rasterMillis = [];

  void onTimings(List<FrameTiming> timings) {
    for (final timing in timings) {
      uiMillis.add(timing.buildDuration.inMicroseconds / 1000);
      rasterMillis.add(timing.rasterDuration.inMicroseconds / 1000);
    }
  }

  bool get isEmpty => uiMillis.isEmpty;

  static double percentile(List<double> values, double fraction) {
    final sorted = [...values]..sort();
    final index = ((sorted.length - 1) * fraction).round();
    return sorted[index];
  }

  Map<String, dynamic> summarise() {
    return {
      'frame_count': uiMillis.length,
      'frame_budget_ms': frameBudgetMs,
      'ui_thread_ms': {
        'p50': percentile(uiMillis, 0.50),
        'p90': percentile(uiMillis, 0.90),
        'p99': percentile(uiMillis, 0.99),
        'worst': uiMillis.reduce((a, b) => a > b ? a : b),
      },
      'raster_thread_ms': {
        'p50': percentile(rasterMillis, 0.50),
        'p90': percentile(rasterMillis, 0.90),
        'p99': percentile(rasterMillis, 0.99),
        'worst': rasterMillis.reduce((a, b) => a > b ? a : b),
      },
      'frames_over_budget': uiMillis.where((v) => v > frameBudgetMs).length,
    };
  }
}

List<StreamModel> _fakeStreams(int count, int tick) {
  return List.generate(count, (i) {
    return StreamModel(
      id: 'stream-$i',
      title: 'Live session $i',
      broadcasterId: 'broadcaster-$i',
      broadcasterName: 'Diffuseur $i',
      // The listener count is what actually changes on every tick, so each
      // rebuild produces a genuinely different tree rather than a no-op.
      listenerCount: (tick + i) % 500,
      description: 'Flux audio en direct numero $i',
      streamUrl: 'https://example.invalid/streams/$i/audio',
      isLive: i.isEven,
      createdAt: DateTime(2026, 1, 1).add(Duration(minutes: i)),
    );
  });
}

/// Rebuilds the list on a ticker to emulate a continuous data feed.
class _LiveList extends StatefulWidget {
  const _LiveList();

  @override
  State<_LiveList> createState() => _LiveListState();
}

class _LiveListState extends State<_LiveList> {
  int _tick = 0;
  bool _running = true;

  @override
  void initState() {
    super.initState();
    _pump();
  }

  Future<void> _pump() async {
    while (_running) {
      await Future<void>.delayed(dataTickInterval);
      if (!mounted) return;
      setState(() => _tick++);
    }
  }

  @override
  void dispose() {
    _running = false;
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final streams = _fakeStreams(60, _tick);
    return MaterialApp(
      home: Scaffold(
        body: ListView.builder(
          itemCount: streams.length,
          itemBuilder: (_, i) => Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            child: StreamCard(stream: streams[i], onTap: () {}),
          ),
        ),
      ),
    );
  }
}

void main() {
  final binding = IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets(
    'the stream list holds the 60 FPS budget while data keeps updating',
    (tester) async {
      final recorder = _FrameRecorder();
      SchedulerBinding.instance.addTimingsCallback(recorder.onTimings);
      addTearDown(
        () =>
            SchedulerBinding.instance.removeTimingsCallback(recorder.onTimings),
      );

      await tester.pumpWidget(const _LiveList());
      await tester.pumpAndSettle(const Duration(milliseconds: 100));

      // Scroll the list while the ticker keeps rebuilding it: this is the
      // "rendering under a real-time feed" case the brief asks to prove.
      for (var pass = 0; pass < 6; pass++) {
        await tester.fling(find.byType(ListView), const Offset(0, -400), 2000);
        await tester.pump(const Duration(milliseconds: 16));
        await tester.pump(const Duration(milliseconds: 400));
        await tester.fling(find.byType(ListView), const Offset(0, 400), 2000);
        await tester.pump(const Duration(milliseconds: 16));
        await tester.pump(const Duration(milliseconds: 400));
      }

      expect(
        recorder.isEmpty,
        isFalse,
        reason:
            'no frame timing was captured: run this on a device, not in a '
            'plain widget test where the clock is faked',
      );

      final summary = recorder.summarise();
      final report = const JsonEncoder.withIndent('  ').convert(summary);
      debugPrint(
        'FRAME_TIMING_SUMMARY_BEGIN\n$report\nFRAME_TIMING_SUMMARY_END',
      );

      // The driver collects this on the host as
      // build/integration_response_data.json — writing a file here would land in
      // the application sandbox on the device instead.
      binding.reportData = <String, dynamic>{'frame_timings': summary};

      // The p90 is the meaningful gate: an isolated slow frame during a fling is
      // expected, a sustained overrun is what a user perceives as jank.
      final uiP90 = (summary['ui_thread_ms'] as Map)['p90'] as double;
      expect(
        uiP90,
        lessThan(frameBudgetMs),
        reason:
            'the UI thread exceeded the 16.67 ms budget on more than 10% '
            'of frames, which is visible stutter',
      );
    },
  );
}
