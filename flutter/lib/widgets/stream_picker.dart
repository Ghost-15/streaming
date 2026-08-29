import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/stream_model.dart';
import '../notifiers/stream_notifier.dart';

/// Lets the user choose a live stream instead of typing its name.
///
/// Favourites and playlist items are keyed on a stream identifier, so a
/// free-text field can only ever produce a value the API rejects. [onPick]
/// reports whether the action succeeded; a failure is surfaced as a snack bar
/// rather than being swallowed.
Future<void> showStreamPicker(
  BuildContext context, {
  required String title,
  required Future<bool> Function(StreamModel stream) onPick,
  required String failureLabel,
}) {
  return showDialog<void>(
    context: context,
    builder: (_) => _StreamPickerDialog(
      title: title,
      onPick: onPick,
      failureLabel: failureLabel,
    ),
  );
}

class _StreamPickerDialog extends StatefulWidget {
  final String title;
  final Future<bool> Function(StreamModel stream) onPick;
  final String failureLabel;

  const _StreamPickerDialog({
    required this.title,
    required this.onPick,
    required this.failureLabel,
  });

  @override
  State<_StreamPickerDialog> createState() => _StreamPickerDialogState();
}

class _StreamPickerDialogState extends State<_StreamPickerDialog> {
  bool _submitting = false;

  @override
  void initState() {
    super.initState();
    // The library screen does not load streams on its own, so make sure the
    // picker has something to show when it is opened from there.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final streams = context.read<StreamNotifier>();
      if (streams.streams.isEmpty && !streams.isLoading) {
        streams.loadActive();
      }
    });
  }

  Future<void> _pick(StreamModel stream) async {
    if (_submitting) return;
    setState(() => _submitting = true);

    final messenger = ScaffoldMessenger.of(context);
    final navigator = Navigator.of(context);
    final ok = await widget.onPick(stream);

    if (!mounted) return;
    navigator.pop();
    if (!ok) {
      messenger.showSnackBar(
        SnackBar(
          content: Text(widget.failureLabel),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final streams = context.watch<StreamNotifier>();
    final cs = Theme.of(context).colorScheme;

    return AlertDialog(
      title: Text(widget.title),
      content: SizedBox(width: double.maxFinite, child: _body(streams, cs)),
      actions: [
        TextButton(
          onPressed: _submitting ? null : () => Navigator.pop(context),
          child: const Text('Annuler'),
        ),
      ],
    );
  }

  Widget _body(StreamNotifier streams, ColorScheme cs) {
    if (streams.isLoading && streams.streams.isEmpty) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 24),
        child: Center(child: CircularProgressIndicator()),
      );
    }

    if (streams.error != null && streams.streams.isEmpty) {
      return Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Impossible de charger les directs.',
            style: TextStyle(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 12),
          FilledButton.tonal(
            onPressed: () => context.read<StreamNotifier>().loadActive(),
            child: const Text('Réessayer'),
          ),
        ],
      );
    }

    if (streams.streams.isEmpty) {
      return Text(
        'Aucun direct en cours.\nReviens quand un diffuseur sera à l’antenne.',
        style: TextStyle(color: cs.onSurfaceVariant),
      );
    }

    return ListView.builder(
      shrinkWrap: true,
      itemCount: streams.streams.length,
      itemBuilder: (_, i) {
        final stream = streams.streams[i];
        return ListTile(
          enabled: !_submitting,
          leading: Icon(
            stream.isLive ? Icons.podcasts_rounded : Icons.radio_rounded,
            color: stream.isLive ? cs.error : cs.primary,
          ),
          title: Text(stream.title, maxLines: 1, overflow: TextOverflow.ellipsis),
          subtitle: Text(
            stream.broadcasterName.isEmpty
                ? 'Diffuseur inconnu'
                : stream.broadcasterName,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
          onTap: () => _pick(stream),
        );
      },
    );
  }
}
