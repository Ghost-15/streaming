import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/playlist_model.dart';
import '../api/models/track_model.dart';
import '../notifiers/favorite_notifier.dart';
import '../notifiers/playlist_notifier.dart';
import '../widgets/loading_indicator.dart';

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
    });
  }

  @override
  void dispose() {
    _tabs.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Ma bibliotheque'),
        bottom: TabBar(
          controller: _tabs,
          tabs: const [
            Tab(icon: Icon(Icons.queue_music), text: 'Playlists'),
            Tab(icon: Icon(Icons.favorite), text: 'Favoris'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabs,
        children: const [_PlaylistsTab(), _FavoritesTab()],
      ),
    );
  }
}

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
          empty: 'Aucune playlist',
          onRetry: () => context.read<PlaylistNotifier>().load(),
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: playlists.playlists.length,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
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
        icon: const Icon(Icons.add),
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
    return Card(
      child: ExpansionTile(
        title: Text(playlist.title),
        subtitle: Text('${playlist.trackCount} titre(s)'),
        onExpansionChanged: (open) {
          if (open) context.read<PlaylistNotifier>().open(playlist.id);
        },
        trailing: PopupMenuButton<_PlaylistAction>(
          onSelected: (action) => _handleAction(context, action),
          itemBuilder: (_) => const [
            PopupMenuItem(
              value: _PlaylistAction.rename,
              child: Text('Renommer'),
            ),
            PopupMenuItem(
              value: _PlaylistAction.addTrack,
              child: Text('Ajouter un titre'),
            ),
            PopupMenuItem(
              value: _PlaylistAction.next,
              child: Text('Titre suivant'),
            ),
            PopupMenuItem(
              value: _PlaylistAction.delete,
              child: Text('Supprimer'),
            ),
          ],
        ),
        children: [
          Consumer<PlaylistNotifier>(
            builder: (context, notifier, _) {
              final selected = notifier.selected?.id == playlist.id
                  ? notifier.selected
                  : playlist;
              if (selected!.tracks.isEmpty) {
                return const Padding(
                  padding: EdgeInsets.all(16),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text('Aucun titre dans cette playlist'),
                  ),
                );
              }
              return Column(
                children: selected.tracks
                    .map(
                      (track) => _TrackTile(
                        track: track,
                        trailing: IconButton(
                          icon: const Icon(Icons.remove_circle_outline),
                          tooltip: 'Retirer',
                          onPressed: () => context
                              .read<PlaylistNotifier>()
                              .removeTrack(playlist.id, track.id),
                        ),
                      ),
                    )
                    .toList(),
              );
            },
          ),
        ],
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
        _showTextDialog(
          context,
          title: 'Ajouter un titre',
          label: 'ID du titre',
          onSubmit: (value) => notifier.addTrack(playlist.id, value),
        );
      case _PlaylistAction.next:
        notifier.next(playlist.id);
      case _PlaylistAction.delete:
        notifier.delete(playlist.id);
    }
  }
}

enum _PlaylistAction { rename, addTrack, next, delete }

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
          empty: 'Aucun favori',
          onRetry: () => context.read<FavoriteNotifier>().load(),
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: favorites.tracks.length,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemBuilder: (_, i) {
              final track = favorites.tracks[i];
              return _TrackTile(
                track: track,
                trailing: IconButton(
                  icon: const Icon(Icons.favorite),
                  tooltip: 'Retirer des favoris',
                  onPressed: () =>
                      context.read<FavoriteNotifier>().remove(track.id),
                ),
              );
            },
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showTextDialog(
          context,
          title: 'Ajouter un favori',
          label: 'ID du titre',
          onSubmit: (value) => context.read<FavoriteNotifier>().add(value),
        ),
        icon: const Icon(Icons.favorite_border),
        label: const Text('Favori'),
      ),
    );
  }
}

class _TrackTile extends StatelessWidget {
  final TrackModel track;
  final Widget? trailing;

  const _TrackTile({required this.track, this.trailing});

  @override
  Widget build(BuildContext context) {
    final subtitle = [
      if (track.artist.isNotEmpty) track.artist,
      if (track.duration > 0) '${track.duration}s',
    ].join(' - ');

    return ListTile(
      leading: const Icon(Icons.music_note),
      title: Text(track.title),
      subtitle: subtitle.isEmpty ? Text(track.id) : Text(subtitle),
      trailing: trailing,
    );
  }
}

class _StatusView extends StatelessWidget {
  final LibraryStatus status;
  final String error;
  final String empty;
  final VoidCallback onRetry;
  final Widget child;

  const _StatusView({
    required this.status,
    required this.error,
    required this.empty,
    required this.onRetry,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    if (status == LibraryStatus.loading) {
      return const LoadingIndicator(message: 'Chargement...');
    }
    if (status == LibraryStatus.error) {
      return ListView(
        padding: const EdgeInsets.all(24),
        children: [
          Icon(
            Icons.error_outline,
            size: 48,
            color: Theme.of(context).colorScheme.error,
          ),
          const SizedBox(height: 16),
          Text(error, textAlign: TextAlign.center),
          const SizedBox(height: 16),
          FilledButton.icon(
            onPressed: onRetry,
            icon: const Icon(Icons.refresh),
            label: const Text('Reessayer'),
          ),
        ],
      );
    }
    if (_isEmptyChild(child)) {
      return ListView(
        padding: const EdgeInsets.all(24),
        children: [Center(child: Text(empty))],
      );
    }
    return child;
  }

  bool _isEmptyChild(Widget widget) {
    if (widget is ListView) {
      final delegate = widget.childrenDelegate;
      if (delegate is SliverChildBuilderDelegate) {
        return delegate.estimatedChildCount == 0;
      }
    }
    return false;
  }
}

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
