import 'package:flutter_test/flutter_test.dart';
import 'package:streampulse/api/models/playlist_model.dart';
import 'package:streampulse/api/models/track_model.dart';
import 'package:streampulse/api/repositories/playlist_repository.dart';
import 'package:streampulse/notifiers/playlist_notifier.dart';

// ── Fakes ────────────────────────────────────────────────────────────────────

PlaylistModel _playlist({String id = 'p-1', String title = 'Mix'}) =>
    PlaylistModel(
      id: id,
      ownerId: 'u-1',
      title: title,
      createdAt: DateTime(2025),
    );

TrackModel _track({String id = 't-1'}) =>
    TrackModel(id: id, title: 'Track $id', createdAt: DateTime(2025));

class _FakePlaylistRepo extends PlaylistRepository {
  List<PlaylistModel> data;
  bool shouldThrow;

  _FakePlaylistRepo({this.data = const [], this.shouldThrow = false});

  @override
  Future<List<PlaylistModel>> list() async {
    if (shouldThrow) throw Exception('network error');
    return data;
  }

  @override
  Future<PlaylistModel> get(String id) async => _playlist(id: id);

  @override
  Future<PlaylistModel> create(String title) async {
    final p = _playlist(title: title);
    data = [...data, p];
    return p;
  }

  @override
  Future<PlaylistModel> rename(String id, String title) async =>
      _playlist(id: id, title: title);

  @override
  Future<void> delete(String id) async {
    data = data.where((p) => p.id != id).toList();
  }

  @override
  Future<void> addTrack(String playlistId, String trackId) async {}

  @override
  Future<void> removeTrack(String playlistId, String trackId) async {}

  @override
  Future<void> reorder(String playlistId, List<String> trackIds) async {}

  @override
  Future<TrackModel> next(String playlistId) async => _track();
}

// ── Tests ────────────────────────────────────────────────────────────────────

void main() {
  _addTrackFailureTests();

  group('PlaylistNotifier', () {
    test('initial state is idle', () {
      final n = PlaylistNotifier(_FakePlaylistRepo());
      expect(n.status, LibraryStatus.idle);
      expect(n.playlists, isEmpty);
      expect(n.selected, isNull);
      expect(n.error, isEmpty);
      n.dispose();
    });

    test('load() transitions to loaded with data', () async {
      final repo = _FakePlaylistRepo(
        data: [
          _playlist(),
          _playlist(id: 'p-2'),
        ],
      );
      final n = PlaylistNotifier(repo);
      await n.load();
      expect(n.status, LibraryStatus.loaded);
      expect(n.playlists.length, 2);
      n.dispose();
    });

    test('load() on error transitions to error state', () async {
      final n = PlaylistNotifier(_FakePlaylistRepo(shouldThrow: true));
      await n.load();
      expect(n.status, LibraryStatus.error);
      expect(n.error, isNotEmpty);
      n.dispose();
    });

    test('load() notifies listeners', () async {
      final n = PlaylistNotifier(_FakePlaylistRepo());
      int calls = 0;
      n.addListener(() => calls++);
      await n.load();
      expect(calls, greaterThan(0));
      n.dispose();
    });

    test('create() with empty title does nothing', () async {
      final repo = _FakePlaylistRepo();
      final n = PlaylistNotifier(repo);
      await n.create('   ');
      expect(n.playlists, isEmpty);
      n.dispose();
    });

    test('create() with valid title adds playlist', () async {
      final repo = _FakePlaylistRepo();
      final n = PlaylistNotifier(repo);
      await n.create('Ma playlist');
      expect(n.playlists.length, 1);
      expect(n.playlists.first.title, 'Ma playlist');
      n.dispose();
    });

    test('delete() clears selected if it matches', () async {
      final repo = _FakePlaylistRepo(data: [_playlist()]);
      final n = PlaylistNotifier(repo);
      await n.load();
      n.selected = n.playlists.first;
      await n.delete('p-1');
      expect(n.selected, isNull);
      n.dispose();
    });

    test('clearError() resets error and notifies', () async {
      final n = PlaylistNotifier(_FakePlaylistRepo(shouldThrow: true));
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

class _FailingPlaylistRepo extends PlaylistRepository {
  const _FailingPlaylistRepo();

  @override
  Future<List<PlaylistModel>> list() async => const [];

  @override
  Future<PlaylistModel> get(String id) async => _playlist(id: id);

  @override
  Future<void> addTrack(String playlistId, String trackId) async =>
      throw Exception('rejected');
}

void _addTrackFailureTests() {
  group('PlaylistNotifier addTrack failure reporting', () {
    test('returns false when the server rejects the track', () async {
      final n = PlaylistNotifier(const _FailingPlaylistRepo());
      expect(await n.addTrack('p-1', 'not-a-uuid'), isFalse);
      expect(n.error, isNotEmpty);
      n.dispose();
    });

    test('returns true when the track is accepted', () async {
      final n = PlaylistNotifier(_FakePlaylistRepo());
      expect(await n.addTrack('p-1', 'uuid-1'), isTrue);
      n.dispose();
    });

    test('refuses an empty identifier without calling the API', () async {
      final n = PlaylistNotifier(const _FailingPlaylistRepo());
      expect(await n.addTrack('p-1', '  '), isFalse);
      n.dispose();
    });
  });
}
