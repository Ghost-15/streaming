import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/api/models/track_model.dart';
import 'package:streampulse/api/repositories/favorite_repository.dart';
import 'package:streampulse/notifiers/favorite_notifier.dart';
import 'package:streampulse/notifiers/playlist_notifier.dart';

// ── Fake ─────────────────────────────────────────────────────────────────────

TrackModel _track(String id) =>
    TrackModel(id: id, title: 'Track $id', createdAt: DateTime(2025));

class _FakeFavoriteRepo extends FavoriteRepository {
  List<TrackModel> data;
  bool shouldThrow;

  _FakeFavoriteRepo({this.data = const [], this.shouldThrow = false});

  @override
  Future<List<TrackModel>> list() async {
    if (shouldThrow) throw Exception('network error');
    return data;
  }

  @override
  Future<void> add(String trackId) async {
    data = [...data, _track(trackId)];
  }

  @override
  Future<void> remove(String trackId) async {
    data = data.where((t) => t.id != trackId).toList();
  }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('FavoriteNotifier', () {
    test('initial state is idle with no tracks', () {
      final n = FavoriteNotifier(_FakeFavoriteRepo());
      expect(n.status, LibraryStatus.idle);
      expect(n.tracks, isEmpty);
      expect(n.error, isEmpty);
      n.dispose();
    });

    test('load() transitions to loaded with tracks', () async {
      final repo = _FakeFavoriteRepo(data: [_track('t-1'), _track('t-2')]);
      final n = FavoriteNotifier(repo);
      await n.load();
      expect(n.status, LibraryStatus.loaded);
      expect(n.tracks.length, 2);
      n.dispose();
    });

    test('load() on error sets error state', () async {
      final n = FavoriteNotifier(_FakeFavoriteRepo(shouldThrow: true));
      await n.load();
      expect(n.status, LibraryStatus.error);
      expect(n.error, isNotEmpty);
      n.dispose();
    });

    test('load() notifies listeners', () async {
      final n = FavoriteNotifier(_FakeFavoriteRepo());
      int calls = 0;
      n.addListener(() => calls++);
      await n.load();
      expect(calls, greaterThan(0));
      n.dispose();
    });

    test('add() with empty id does nothing', () async {
      final repo = _FakeFavoriteRepo();
      final n = FavoriteNotifier(repo);
      await n.add('   ');
      expect(n.tracks, isEmpty);
      n.dispose();
    });

    test('add() with valid id adds track and reloads', () async {
      final repo = _FakeFavoriteRepo();
      final n = FavoriteNotifier(repo);
      await n.add('t-42');
      expect(n.tracks.any((t) => t.id == 't-42'), isTrue);
      n.dispose();
    });

    test('remove() removes track and reloads', () async {
      final repo = _FakeFavoriteRepo(data: [_track('t-1'), _track('t-2')]);
      final n = FavoriteNotifier(repo);
      await n.load();
      await n.remove('t-1');
      expect(n.tracks.any((t) => t.id == 't-1'), isFalse);
      n.dispose();
    });

    test('clearError() resets error and notifies', () async {
      final n = FavoriteNotifier(_FakeFavoriteRepo(shouldThrow: true));
      await n.load();
      bool notified = false;
      n.addListener(() => notified = true);
      n.clearError();
      expect(n.error, isEmpty);
      expect(notified, isTrue);
      n.dispose();
    });
  });
}
