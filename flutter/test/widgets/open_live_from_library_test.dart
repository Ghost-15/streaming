import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';

import 'package:streampulse/api/models/playlist_model.dart';
import 'package:streampulse/api/models/stream_model.dart';
import 'package:streampulse/api/models/track_model.dart';
import 'package:streampulse/api/repositories/favorite_repository.dart';
import 'package:streampulse/api/repositories/playlist_repository.dart';
import 'package:streampulse/api/repositories/stream_repository.dart';
import 'package:streampulse/notifiers/audio_notifier.dart';
import 'package:streampulse/notifiers/favorite_notifier.dart';
import 'package:streampulse/notifiers/playlist_notifier.dart';
import 'package:streampulse/notifiers/stream_notifier.dart';
import 'package:streampulse/screens/library_screen.dart';

// A favourite stores only a stream identifier. Tapping it must open the live,
// but only while the broadcaster is actually on air: GET /streams returns rows
// with status 'live', so absence from that list means the broadcast has ended.

const _liveId = 'stream-live';
const _endedId = 'stream-ended';

StreamModel _live() => StreamModel(
  id: _liveId,
  title: 'Radio Nuit',
  broadcasterId: 'b-1',
  broadcasterName: 'Diffuseur',
  streamUrl: 'https://example.invalid/live',
  isLive: true,
  createdAt: DateTime(2026),
);

TrackModel _entry(String id, String title) =>
    TrackModel(id: id, title: title, createdAt: DateTime(2026));

class _FakeStreamRepo extends StreamRepository {
  final List<StreamModel> data;
  const _FakeStreamRepo(this.data);

  @override
  Future<List<StreamModel>> getActive() async => data;
}

class _FakeFavoriteRepo extends FavoriteRepository {
  final List<TrackModel> data;
  const _FakeFavoriteRepo(this.data);

  @override
  Future<List<TrackModel>> list() async => data;
}

class _FakePlaylistRepo extends PlaylistRepository {
  const _FakePlaylistRepo();

  @override
  Future<List<PlaylistModel>> list() async => const [];
}

/// Records what the library asked the player to do.
class _SpyAudioNotifier extends AudioNotifier {
  final List<String> played = [];
  StreamModel? _current;

  @override
  StreamModel? get currentStream => _current;

  void pretendPlaying(StreamModel stream) => _current = stream;

  @override
  Future<void> playStream(StreamModel stream) async {
    played.add(stream.id);
    _current = stream;
  }
}

Widget _host({
  required _SpyAudioNotifier audio,
  required List<StreamModel> live,
  required List<TrackModel> favourites,
}) {
  return MultiProvider(
    providers: [
      ChangeNotifierProvider<AudioNotifier>.value(value: audio),
      ChangeNotifierProvider(
        create: (_) => StreamNotifier(_FakeStreamRepo(live)),
      ),
      ChangeNotifierProvider(
        create: (_) => FavoriteNotifier(_FakeFavoriteRepo(favourites)),
      ),
      ChangeNotifierProvider(
        create: (_) => PlaylistNotifier(const _FakePlaylistRepo()),
      ),
    ],
    child: const MaterialApp(home: LibraryScreen()),
  );
}

Future<void> _openFavouritesTab(WidgetTester tester) async {
  await tester.pumpAndSettle();
  await tester.tap(find.text('Favoris'));
  await tester.pumpAndSettle();
}

void main() {
  _badgeTests();

  testWidgets('tapping a favourite whose broadcast is on air opens it', (
    tester,
  ) async {
    final audio = _SpyAudioNotifier();
    await tester.pumpWidget(
      _host(
        audio: audio,
        live: [_live()],
        favourites: [_entry(_liveId, 'Radio Nuit')],
      ),
    );
    await _openFavouritesTab(tester);

    await tester.tap(find.text('Radio Nuit'));
    await tester.pumpAndSettle();

    expect(audio.played, [_liveId]);
  });

  testWidgets('tapping a favourite whose broadcast ended explains why', (
    tester,
  ) async {
    final audio = _SpyAudioNotifier();
    await tester.pumpWidget(
      _host(
        audio: audio,
        live: const [], // nothing on air
        favourites: [_entry(_endedId, 'Emission finie')],
      ),
    );
    await _openFavouritesTab(tester);

    await tester.tap(find.text('Emission finie'));
    await tester.pumpAndSettle();

    expect(
      audio.played,
      isEmpty,
      reason: 'a stream that is not on air must not be started',
    );
    expect(find.textContaining('n’est pas à l’antenne'), findsOneWidget);
  });

  testWidgets('tapping the live already playing does not restart it', (
    tester,
  ) async {
    final audio = _SpyAudioNotifier()..pretendPlaying(_live());
    await tester.pumpWidget(
      _host(
        audio: audio,
        live: [_live()],
        favourites: [_entry(_liveId, 'Radio Nuit')],
      ),
    );
    await _openFavouritesTab(tester);

    await tester.tap(find.text('Radio Nuit'));
    await tester.pumpAndSettle();

    expect(
      audio.played,
      isEmpty,
      reason: 'restarting the running stream would cut the listener off',
    );
  });
}

// ── Live / offline badge ─────────────────────────────────────────────────────

void _badgeTests() {
  testWidgets('a favourite on air is labelled as live', (tester) async {
    await tester.pumpWidget(
      _host(
        audio: _SpyAudioNotifier(),
        live: [_live()],
        favourites: [_entry(_liveId, 'Radio Nuit')],
      ),
    );
    await _openFavouritesTab(tester);

    expect(find.textContaining('En direct'), findsOneWidget);
    expect(find.textContaining('Hors ligne'), findsNothing);
  });

  testWidgets('a favourite whose broadcast ended is labelled offline', (
    tester,
  ) async {
    await tester.pumpWidget(
      _host(
        audio: _SpyAudioNotifier(),
        live: const [],
        favourites: [_entry(_endedId, 'Emission finie')],
      ),
    );
    await _openFavouritesTab(tester);

    expect(find.textContaining('Hors ligne'), findsOneWidget);
    expect(find.textContaining('En direct'), findsNothing);
  });

  testWidgets('the state is written, not only coloured', (tester) async {
    // Colour alone excludes screen reader users and anyone who cannot
    // distinguish the dot, so the wording must exist as text.
    await tester.pumpWidget(
      _host(
        audio: _SpyAudioNotifier(),
        live: [_live()],
        favourites: [_entry(_liveId, 'Radio Nuit')],
      ),
    );
    await _openFavouritesTab(tester);

    expect(find.bySemanticsLabel(RegExp('En direct')), findsWidgets);
  });
}
