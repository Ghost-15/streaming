import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';

import 'package:streampulse/api/models/stream_model.dart';
import 'package:streampulse/api/repositories/stream_repository.dart';
import 'package:streampulse/notifiers/stream_notifier.dart';
import 'package:streampulse/widgets/stream_picker.dart';

// Favourites and playlist items are keyed on a stream identifier. These tests
// lock in that the picker hands back that identifier — the previous free-text
// dialog sent whatever the user typed, which the API could only reject.

StreamModel _stream(String id, String title) => StreamModel(
  id: id,
  title: title,
  broadcasterId: 'b-$id',
  broadcasterName: 'Diffuseur $id',
  streamUrl: 'https://example.invalid/$id',
  isLive: true,
  createdAt: DateTime(2026),
);

class _FakeStreamRepo extends StreamRepository {
  final List<StreamModel> data;
  final bool shouldThrow;

  const _FakeStreamRepo({this.data = const [], this.shouldThrow = false});

  @override
  Future<List<StreamModel>> getActive() async {
    if (shouldThrow) throw Exception('network down');
    return data;
  }
}

Widget _host({
  required StreamNotifier notifier,
  required Future<bool> Function(StreamModel) onPick,
}) {
  return ChangeNotifierProvider<StreamNotifier>.value(
    value: notifier,
    child: MaterialApp(
      home: Scaffold(
        body: Builder(
          builder: (context) => ElevatedButton(
            onPressed: () => showStreamPicker(
              context,
              title: 'Ajouter un favori',
              onPick: onPick,
              failureLabel: 'Ajout impossible',
            ),
            child: const Text('ouvrir'),
          ),
        ),
      ),
    ),
  );
}

void main() {
  testWidgets('picking a stream hands back its identifier, not its title', (
    tester,
  ) async {
    final notifier = StreamNotifier(
      _FakeStreamRepo(data: [_stream('uuid-1', 'Radio Nuit')]),
    );
    String? received;

    await tester.pumpWidget(
      _host(
        notifier: notifier,
        onPick: (s) async {
          received = s.id;
          return true;
        },
      ),
    );
    await tester.tap(find.text('ouvrir'));
    await tester.pumpAndSettle();

    expect(find.text('Radio Nuit'), findsOneWidget);
    await tester.tap(find.text('Radio Nuit'));
    await tester.pumpAndSettle();

    expect(
      received,
      'uuid-1',
      reason: 'the API needs the stream id; sending the title always fails',
    );
    notifier.dispose();
  });

  testWidgets('a failed add is surfaced instead of being swallowed', (
    tester,
  ) async {
    final notifier = StreamNotifier(
      _FakeStreamRepo(data: [_stream('uuid-1', 'Radio Nuit')]),
    );

    await tester.pumpWidget(
      _host(notifier: notifier, onPick: (_) async => false),
    );
    await tester.tap(find.text('ouvrir'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Radio Nuit'));
    await tester.pumpAndSettle();

    expect(
      find.text('Ajout impossible'),
      findsOneWidget,
      reason: 'the user must be told the action did not work',
    );
    notifier.dispose();
  });

  testWidgets('a successful add closes the dialog without an alert', (
    tester,
  ) async {
    final notifier = StreamNotifier(
      _FakeStreamRepo(data: [_stream('uuid-1', 'Radio Nuit')]),
    );

    await tester.pumpWidget(
      _host(notifier: notifier, onPick: (_) async => true),
    );
    await tester.tap(find.text('ouvrir'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Radio Nuit'));
    await tester.pumpAndSettle();

    expect(find.text('Ajout impossible'), findsNothing);
    expect(find.text('Radio Nuit'), findsNothing);
    notifier.dispose();
  });

  testWidgets('an empty list explains why nothing can be picked', (
    tester,
  ) async {
    final notifier = StreamNotifier(const _FakeStreamRepo());

    await tester.pumpWidget(
      _host(notifier: notifier, onPick: (_) async => true),
    );
    await tester.tap(find.text('ouvrir'));
    await tester.pumpAndSettle();

    expect(find.textContaining('Aucun direct en cours'), findsOneWidget);
    notifier.dispose();
  });

  testWidgets('a load failure offers a retry', (tester) async {
    final notifier = StreamNotifier(const _FakeStreamRepo(shouldThrow: true));

    await tester.pumpWidget(
      _host(notifier: notifier, onPick: (_) async => true),
    );
    await tester.tap(find.text('ouvrir'));
    await tester.pumpAndSettle();

    expect(find.textContaining('Impossible de charger'), findsOneWidget);
    expect(find.text('Réessayer'), findsOneWidget);
    notifier.dispose();
  });
}
