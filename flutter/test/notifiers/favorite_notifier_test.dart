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
  Future<void> add(String streamId) async {
    data = [...data, _track(streamId)];
  }

  @override
  Future<void> remove(String streamId) async {
    data = data.where((t) => t.id != streamId).toList();
  }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  _failureReportingTests();

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

// ── Failure reporting ────────────────────────────────────────────────────────

class _FailingFavoriteRepo extends FavoriteRepository {
  const _FailingFavoriteRepo();

  @override
  Future<List<TrackModel>> list() async => const [];

  @override
  Future<void> add(String streamId) async => throw Exception('rejected');

  @override
  Future<void> remove(String streamId) async => throw Exception('rejected');
}

void _failureReportingTests() {
  group('FavoriteNotifier failure reporting', () {
    test('add returns false when the server rejects the stream', () async {
      final n = FavoriteNotifier(const _FailingFavoriteRepo());
      expect(await n.add('not-a-uuid'), isFalse);
      expect(n.error, isNotEmpty);
      n.dispose();
    });

    test('add returns true when the stream is accepted', () async {
      final n = FavoriteNotifier(_FakeFavoriteRepo());
      expect(await n.add('uuid-1'), isTrue);
      n.dispose();
    });

    test('add refuses an empty identifier without calling the API', () async {
      final n = FavoriteNotifier(const _FailingFavoriteRepo());
      expect(await n.add('   '), isFalse);
      n.dispose();
    });

    test('remove returns false when the server rejects it', () async {
      final n = FavoriteNotifier(const _FailingFavoriteRepo());
      expect(await n.remove('uuid-1'), isFalse);
      n.dispose();
    });
  });
}
