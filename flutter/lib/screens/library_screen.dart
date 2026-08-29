import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/playlist_model.dart';
import '../api/models/stream_model.dart';
import '../api/models/track_model.dart';
import '../notifiers/audio_notifier.dart';
import '../notifiers/favorite_notifier.dart';
import '../notifiers/playlist_notifier.dart';
import '../notifiers/stream_notifier.dart';
import '../widgets/loading_indicator.dart';
import '../widgets/page_header.dart';
import '../widgets/stream_picker.dart';

class LibraryScreen extends StatefulWidget {
  const LibraryScreen({super.key});

  @override
  State<LibraryScreen> createState() => _LibraryScreenState();
}

class _LibraryScreenState extends State<LibraryScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabs;

  @override
  void initState() {
    super.initState();
    _tabs = TabController(length: 2, vsync: this);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      context.read<PlaylistNotifier>().load();
      context.read<FavoriteNotifier>().load();
      // Needed to resolve a favourite or a playlist item to a live stream.
      context.read<StreamNotifier>().loadActive();
    });
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Scaffold(
      body: Column(
        children: [
          const PageHeader(
            icon: Icons.library_music_rounded,
            title: 'Ma Bibliothèque',
            subtitle: 'Playlists et favoris',
          ),
          TabBar(
            controller: _tabs,
            tabs: const [
              Tab(icon: Icon(Icons.queue_music_rounded), text: 'Playlists'),
              Tab(icon: Icon(Icons.favorite_rounded), text: 'Favoris'),
            ],
            indicatorPadding: const EdgeInsets.symmetric(horizontal: 16),
            dividerColor: cs.outlineVariant,
          ),
          Expanded(
            child: TabBarView(
              controller: _tabs,
              children: const [_PlaylistsTab(), _FavoritesTab()],
            ),
          ),
        ],
      ),
    );
  }
}

// ── Playlists tab ─────────────────────────────────────────────────────────────

class _PlaylistsTab extends StatelessWidget {
  const _PlaylistsTab();

  @override
  Widget build(BuildContext context) {
    final playlists = context.watch<PlaylistNotifier>();

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => context.read<PlaylistNotifier>().load(),
        child: _StatusView(
          status: playlists.status,
          error: playlists.error,
          emptyLabel: 'Aucune playlist',
          emptyHint: 'Crée ta première playlist',
          emptyIcon: Icons.queue_music_rounded,
          isEmpty: playlists.playlists.isEmpty,
          onRetry: () => context.read<PlaylistNotifier>().load(),
          child: ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
            itemCount: playlists.playlists.length,
            separatorBuilder: (_, _) => const SizedBox(height: 10),
            itemBuilder: (_, i) =>
                _PlaylistTile(playlist: playlists.playlists[i]),
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showTextDialog(
          context,
          title: 'Nouvelle playlist',
          label: 'Titre',
          onSubmit: (value) => context.read<PlaylistNotifier>().create(value),
        ),
        icon: const Icon(Icons.add_rounded),
        label: const Text('Playlist'),
      ),
    );
  }
}

class _PlaylistTile extends StatelessWidget {
  final PlaylistModel playlist;
  const _PlaylistTile({required this.playlist});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    return Container(
      decoration: BoxDecoration(
        color: cs.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: cs.outlineVariant, width: 0.8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.12),
            blurRadius: 10,
            offset: const Offset(0, 3),
            spreadRadius: -3,
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Theme(
          data: Theme.of(
            context,
          ).copyWith(dividerColor: cs.outlineVariant.withValues(alpha: 0.4)),
          child: ExpansionTile(
            tilePadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 4,
            ),
            childrenPadding: EdgeInsets.zero,
            leading: Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: cs.primaryContainer.withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                Icons.queue_music_rounded,
                size: 20,
                color: cs.primary,
              ),
            ),
            title: Text(
              playlist.title,
              style: tt.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
            ),
            subtitle: Text(
              '${playlist.trackCount} titre${playlist.trackCount != 1 ? 's' : ''}',
              style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
            ),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                PopupMenuButton<_PlaylistAction>(
                  icon: Icon(
                    Icons.more_vert_rounded,
                    size: 18,
                    color: cs.onSurfaceVariant,
                  ),
                  onSelected: (action) => _handleAction(context, action),
                  itemBuilder: (_) => const [
                    PopupMenuItem(
                      value: _PlaylistAction.rename,
                      child: ListTile(
                        leading: Icon(Icons.edit_outlined),
                        title: Text('Renommer'),
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                      ),
                    ),
                    PopupMenuItem(
                      value: _PlaylistAction.addTrack,
                      child: ListTile(
                        leading: Icon(Icons.add_rounded),
                        title: Text('Ajouter un titre'),
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                      ),
                    ),
                    PopupMenuItem(
                      value: _PlaylistAction.next,
                      child: ListTile(
                        leading: Icon(Icons.skip_next_rounded),
                        title: Text('Titre suivant'),
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                      ),
                    ),
                    PopupMenuItem(
                      value: _PlaylistAction.delete,
                      child: ListTile(
                        leading: Icon(Icons.delete_outline_rounded),
                        title: Text('Supprimer'),
                        contentPadding: EdgeInsets.zero,
                        dense: true,
                      ),
                    ),
                  ],
                ),
              ],
            ),
            onExpansionChanged: (open) {
              if (open) context.read<PlaylistNotifier>().open(playlist.id);
            },
            children: [
              Consumer<PlaylistNotifier>(
                builder: (context, notifier, _) {
                  final selected = notifier.selected?.id == playlist.id
                      ? notifier.selected
                      : playlist;
                  if (selected!.tracks.isEmpty) {
                    return Padding(
                      padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                      child: Row(
                        children: [
                          Icon(
                            Icons.music_off_rounded,
                            size: 15,
                            color: cs.onSurfaceVariant,
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'Aucun titre dans cette playlist',
                            style: tt.bodySmall?.copyWith(
                              color: cs.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    );
                  }
                  return Padding(
                    padding: const EdgeInsets.fromLTRB(12, 4, 12, 12),
                    child: Column(
                      children: selected.tracks
                          .map(
                            (track) => Padding(
                              padding: const EdgeInsets.only(bottom: 6),
                              child: _TrackTile(
                                track: track,
                                onTap: () => _openLive(context, track),
                                trailing: IconButton(
                                  icon: Icon(
                                    Icons.remove_circle_outline_rounded,
                                    size: 18,
                                    color: cs.onSurfaceVariant,
                                  ),
                                  tooltip: 'Retirer',
                                  onPressed: () => context
                                      .read<PlaylistNotifier>()
                                      .removeTrack(playlist.id, track.id),
                                ),
                              ),
                            ),
                          )
                          .toList(),
                    ),
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _handleAction(BuildContext context, _PlaylistAction action) {
    final notifier = context.read<PlaylistNotifier>();
    switch (action) {
      case _PlaylistAction.rename:
        _showTextDialog(
          context,
          title: 'Renommer',
          label: 'Titre',
          initialValue: playlist.title,
          onSubmit: (value) => notifier.rename(playlist.id, value),
        );
      case _PlaylistAction.addTrack:
        // The API keys playlist items on a live stream identifier, so the user
        // picks a stream rather than typing a name.
        showStreamPicker(
          context,
          title: 'Ajouter à « ${playlist.title} »',
          onPick: (stream) => notifier.addTrack(playlist.id, stream.id),
          failureLabel: 'Impossible d’ajouter ce direct à la playlist',
        );
      case _PlaylistAction.next:
        notifier.next(playlist.id);
      case _PlaylistAction.delete:
        notifier.delete(playlist.id);
    }
  }
}

enum _PlaylistAction { rename, addTrack, next, delete }

// ── Favorites tab ─────────────────────────────────────────────────────────────

class _FavoritesTab extends StatelessWidget {
  const _FavoritesTab();

  @override
  Widget build(BuildContext context) {
    final favorites = context.watch<FavoriteNotifier>();

    return Scaffold(
      body: RefreshIndicator(
        onRefresh: () => context.read<FavoriteNotifier>().load(),
        child: _StatusView(
          status: favorites.status,
          error: favorites.error,
          emptyLabel: 'Aucun favori',
          emptyHint: 'Ajoute tes titres préférés',
          emptyIcon: Icons.favorite_border_rounded,
          isEmpty: favorites.tracks.isEmpty,
          onRetry: () => context.read<FavoriteNotifier>().load(),
          child: ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 100),
            itemCount: favorites.tracks.length,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemBuilder: (_, i) {
              final track = favorites.tracks[i];
              return _TrackTile(
                track: track,
                onTap: () => _openLive(context, track),
                trailing: IconButton(
                  icon: Icon(
                    Icons.favorite_rounded,
                    size: 18,
                    color: Theme.of(context).colorScheme.error,
                  ),
                  tooltip: 'Retirer des favoris',
                  onPressed: () =>
                      context.read<FavoriteNotifier>().remove(track.id),
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}

// ── Opening a live from the library ───────────────────────────────────────────

StreamModel? _liveWithId(StreamNotifier streams, String id) {
  for (final stream in streams.streams) {
    if (stream.id == id) return stream;
  }
  return null;
}

/// Starts playback of the live a library entry points at.
///
/// A favourite or a playlist item only stores a stream identifier, so the live
/// is looked up among the streams currently on air. `GET /streams` returns only
/// rows with status `live`, which is exactly the "the broadcaster is on air"
/// condition: an entry whose broadcast has ended cannot be opened.
Future<void> _openLive(BuildContext context, TrackModel track) async {
  final audio = context.read<AudioNotifier>();
  final streams = context.read<StreamNotifier>();
  final messenger = ScaffoldMessenger.of(context);

  // Already the current stream: leave the running playback alone.
  if (audio.currentStream?.id == track.id) return;

  var live = _liveWithId(streams, track.id);
  if (live == null) {
    // The cached list may predate this broadcast, so refresh once before
    // telling the user the live is unavailable.
    await streams.loadActive();
    live = _liveWithId(streams, track.id);
  }

  if (live == null) {
    messenger.showSnackBar(
      SnackBar(
        content: Text('« ${track.title} » n’est pas à l’antenne'),
        behavior: SnackBarBehavior.floating,
      ),
    );
    return;
  }

  await audio.playStream(live);
}

// ── Live / offline badge ──────────────────────────────────────────────────────

/// Tells whether the broadcast behind a library entry is currently on air.
///
/// The wording is carried by a real text node rather than by colour alone, so
/// the state is announced to a screen reader and remains readable for someone
/// who does not distinguish the red dot.
class _LiveBadge extends StatelessWidget {
  final bool isLive;
  final String detail;

  const _LiveBadge({required this.isLive, required this.detail});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final color = isLive ? cs.error : cs.onSurfaceVariant;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 7,
          height: 7,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 6),
        Flexible(
          child: Text(
            detail.isEmpty
                ? (isLive ? 'En direct' : 'Hors ligne')
                : '${isLive ? 'En direct' : 'Hors ligne'} · $detail',
            style: tt.bodySmall?.copyWith(
              color: color,
              fontWeight: isLive ? FontWeight.w600 : FontWeight.w400,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}

// ── Track tile ────────────────────────────────────────────────────────────────

class _TrackTile extends StatelessWidget {
  final TrackModel track;
  final Widget? trailing;
  final VoidCallback? onTap;
  const _TrackTile({required this.track, this.trailing, this.onTap});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    // Watched rather than passed in: the badge follows every refresh of the
    // live list without the parent having to rebuild the entry.
    final isLive = context.watch<StreamNotifier>().streams.any(
      (stream) => stream.id == track.id,
    );
    final subtitle = [
      if (track.artist.isNotEmpty) track.artist,
      if (track.duration > 0) '${track.duration}s',
    ].join(' · ');

    return Material(
      color: cs.surfaceContainer,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: cs.outlineVariant.withValues(alpha: 0.5),
              width: 0.6,
            ),
          ),
          child: Row(
            children: [
              Container(
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: isLive
                      ? cs.errorContainer.withValues(alpha: 0.45)
                      : cs.surfaceContainerHighest,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(
                  isLive ? Icons.podcasts_rounded : Icons.radio_rounded,
                  size: 17,
                  color: isLive ? cs.error : cs.onSurfaceVariant,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      track.title,
                      style: tt.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 2),
                    _LiveBadge(isLive: isLive, detail: subtitle),
                  ],
                ),
              ),
              ?trailing,
            ],
          ),
        ),
      ),
    );
  }
}

// ── Status view ───────────────────────────────────────────────────────────────

class _StatusView extends StatelessWidget {
  final LibraryStatus status;
  final String error;
  final String emptyLabel;
  final String emptyHint;
  final IconData emptyIcon;
  final bool isEmpty;
  final VoidCallback onRetry;
  final Widget child;

  const _StatusView({
    required this.status,
    required this.error,
    required this.emptyLabel,
    required this.emptyHint,
    required this.emptyIcon,
    required this.isEmpty,
    required this.onRetry,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;

    if (status == LibraryStatus.loading) {
      return const LoadingIndicator(message: 'Chargement...');
    }

    if (status == LibraryStatus.error) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: cs.errorContainer.withValues(alpha: 0.4),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Icon(
                  Icons.error_outline_rounded,
                  size: 30,
                  color: cs.error,
                ),
              ),
              const SizedBox(height: 16),
              Text(
                error,
                textAlign: TextAlign.center,
                style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
              ),
              const SizedBox(height: 20),
              FilledButton.icon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh_rounded, size: 18),
                label: const Text('Réessayer'),
              ),
            ],
          ),
        ),
      );
    }

    if (isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(40),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 72,
                height: 72,
                decoration: BoxDecoration(
                  color: cs.surfaceContainerHigh,
                  borderRadius: BorderRadius.circular(22),
                  border: Border.all(color: cs.outlineVariant, width: 0.8),
                ),
                child: Icon(emptyIcon, size: 32, color: cs.onSurfaceVariant),
              ),
              const SizedBox(height: 16),
              Text(
                emptyLabel,
                style: tt.titleMedium?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 6),
              Text(
                emptyHint,
                style: tt.bodySmall?.copyWith(color: cs.onSurfaceVariant),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }

    return child;
  }
}

// ── Dialog helper ─────────────────────────────────────────────────────────────

void _showTextDialog(
  BuildContext context, {
  required String title,
  required String label,
  String initialValue = '',
  required ValueChanged<String> onSubmit,
}) {
  final controller = TextEditingController(text: initialValue);
  showDialog<void>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: Text(title),
      content: TextField(
        controller: controller,
        decoration: InputDecoration(labelText: label),
        autofocus: true,
        onSubmitted: (_) {
          onSubmit(controller.text);
          Navigator.pop(ctx);
        },
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(ctx),
          child: const Text('Annuler'),
        ),
        FilledButton(
          onPressed: () {
            onSubmit(controller.text);
            Navigator.pop(ctx);
          },
          child: const Text('Valider'),
        ),
      ],
    ),
  );
}
