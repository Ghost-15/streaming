import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../api/models/stream_model.dart';
import '../notifiers/audio_notifier.dart';
import '../notifiers/broadcaster_notifier.dart';
import '../notifiers/session_notifier.dart';
import '../widgets/page_header.dart';

class BroadcasterScreen extends StatefulWidget {
  const BroadcasterScreen({super.key});

  @override
  State<BroadcasterScreen> createState() => _BroadcasterScreenState();
}

class _BroadcasterScreenState extends State<BroadcasterScreen> {
  final _titleController = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final user = context.read<SessionNotifier>().user;
      if (user != null &&
          (user.role == Role.diffuseur || user.role == Role.admin)) {
        context.read<BroadcasterNotifier>().loadOwned();
      }
    });
  }

  @override
  void dispose() {
    _titleController.dispose();
    super.dispose();
  }

  Future<void> _startLive() async {
    final title = _titleController.text.trim();
    if (title.isNotEmpty) {
      // The mini-player persists in the studio. Stop it before opening the
      // microphone so delayed speaker output cannot be captured and rebroadcast.
      await context.read<AudioNotifier>().stop();
      if (!mounted) return;
    }
    await context.read<BroadcasterNotifier>().startStream(title);
    if (mounted && context.read<BroadcasterNotifier>().isStreaming) {
      _titleController.clear();
    }
  }

  Future<void> _restartLive(StreamModel stream) async {
    await context.read<AudioNotifier>().stop();
    if (!mounted) return;
    await context.read<BroadcasterNotifier>().restartStream(stream);
  }

  Future<void> _deleteLive(StreamModel stream) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Supprimer ce live ?'),
        content: Text(
          '« ${stream.title} » sera supprimé définitivement. Cette action ne peut pas être annulée.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Annuler'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('Supprimer'),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await context.read<BroadcasterNotifier>().deleteStream(stream);
    }
  }

  @override
  Widget build(BuildContext context) {
    final session = context.watch<SessionNotifier>();
    final broadcaster = context.watch<BroadcasterNotifier>();
    final user = session.user;

    if (user == null ||
        (user.role != Role.diffuseur && user.role != Role.admin)) {
      return const _UnauthorizedView();
    }

    return Scaffold(
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const PageHeader(
              icon: Icons.mic_rounded,
              title: 'Mon Studio',
              subtitle: 'Gère et diffuse tes streams',
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  _LiveStatusCard(
                    state: broadcaster.state,
                    streamTitle: broadcaster.currentStream?.title,
                  ),
                  const SizedBox(height: 24),
                  if (!broadcaster.isStreaming) ...[
                    _TitleInput(
                      controller: _titleController,
                      enabled: !broadcaster.isLoading,
                    ),
                    const SizedBox(height: 24),
                  ],
                  if (broadcaster.hasError) ...[
                    _ErrorBanner(
                      message: broadcaster.errorMessage,
                      onDismiss: () =>
                          context.read<BroadcasterNotifier>().clearError(),
                    ),
                    const SizedBox(height: 16),
                  ],
                  if (broadcaster.isLoading)
                    const _LoadingCard()
                  else
                    _ToggleButton(
                      isStreaming: broadcaster.isStreaming,
                      onStart: _startLive,
                      onStop: () =>
                          context.read<BroadcasterNotifier>().stopStream(),
                    ),
                  if (!broadcaster.isStreaming) ...[
                    const SizedBox(height: 32),
                    _OwnedLivesSection(
                      streams: broadcaster.ownedStreams,
                      isLoading: broadcaster.isCatalogLoading,
                      enabled: !broadcaster.isLoading,
                      onRestart: _restartLive,
                      onDelete: _deleteLive,
                    ),
                  ],
                  const SizedBox(height: 120),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _OwnedLivesSection extends StatelessWidget {
  final List<StreamModel> streams;
  final bool isLoading;
  final bool enabled;
  final ValueChanged<StreamModel> onRestart;
  final ValueChanged<StreamModel> onDelete;

  const _OwnedLivesSection({
    required this.streams,
    required this.isLoading,
    required this.enabled,
    required this.onRestart,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          'Mes lives',
          style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 6),
        Text(
          'Reprends un live existant avec le même lien, ou supprime-le définitivement.',
          style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
        ),
        const SizedBox(height: 12),
        if (isLoading)
          const Padding(
            padding: EdgeInsets.all(20),
            child: Center(child: CircularProgressIndicator()),
          )
        else if (streams.isEmpty)
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
              color: cs.surfaceContainerHigh,
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Text('Aucun live enregistré pour le moment.'),
          )
        else
          ...streams.map(
            (stream) => Card(
              margin: const EdgeInsets.only(bottom: 10),
              child: ListTile(
                leading: Icon(
                  stream.isLive ? Icons.podcasts_rounded : Icons.radio_rounded,
                  color: stream.isLive ? cs.error : cs.primary,
                ),
                title: Text(
                  stream.title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                subtitle: Text(stream.isLive ? 'Session active' : 'Hors ligne'),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton(
                      tooltip: 'Continuer ce live',
                      onPressed: enabled ? () => onRestart(stream) : null,
                      icon: const Icon(Icons.play_arrow_rounded),
                    ),
                    IconButton(
                      tooltip: 'Supprimer ce live',
                      onPressed: enabled ? () => onDelete(stream) : null,
                      icon: Icon(Icons.delete_outline_rounded, color: cs.error),
                    ),
                  ],
                ),
              ),
            ),
          ),
      ],
    );
  }
}

// ── Live status card ──────────────────────────────────────────────────────────

class _LiveStatusCard extends StatelessWidget {
  final BroadcasterState state;
  final String? streamTitle;

  const _LiveStatusCard({required this.state, this.streamTitle});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final isLive = state == BroadcasterState.streaming;

    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: isLive
            ? cs.error.withValues(alpha: 0.12)
            : cs.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: isLive ? cs.error.withValues(alpha: 0.4) : cs.outlineVariant,
          width: isLive ? 1.5 : 0.5,
        ),
      ),
      child: Column(
        children: [
          // Status indicator
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isLive)
                _PulsingDot(color: cs.error)
              else
                Container(
                  width: 10,
                  height: 10,
                  decoration: BoxDecoration(
                    color: cs.onSurfaceVariant,
                    borderRadius: BorderRadius.circular(5),
                  ),
                ),
              const SizedBox(width: 10),
              Flexible(
                child: Text(
                  isLive ? 'EN DIRECT' : 'HORS LIGNE',
                  style: tt.titleMedium?.copyWith(
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1.5,
                    color: isLive ? cs.error : cs.onSurfaceVariant,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          if (streamTitle != null && streamTitle!.isNotEmpty) ...[
            const SizedBox(height: 12),
            Text(
              streamTitle!,
              style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w700),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ],
      ),
    );
  }
}

// ── Pulsing dot ───────────────────────────────────────────────────────────────

class _PulsingDot extends StatefulWidget {
  final Color color;
  const _PulsingDot({required this.color});

  @override
  State<_PulsingDot> createState() => _PulsingDotState();
}

class _PulsingDotState extends State<_PulsingDot>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _anim;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);
    _anim = Tween<double>(
      begin: 0.4,
      end: 1.0,
    ).animate(CurvedAnimation(parent: _ctrl, curve: Curves.easeInOut));
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _anim,
      builder: (_, _) => Opacity(
        opacity: _anim.value,
        child: Container(
          width: 10,
          height: 10,
          decoration: BoxDecoration(
            color: widget.color,
            borderRadius: BorderRadius.circular(5),
            boxShadow: [
              BoxShadow(
                color: widget.color.withValues(alpha: 0.5),
                blurRadius: 6,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ── Title input ───────────────────────────────────────────────────────────────

class _TitleInput extends StatelessWidget {
  final TextEditingController controller;
  final bool enabled;
  const _TitleInput({required this.controller, required this.enabled});

  @override
  Widget build(BuildContext context) {
    final tt = Theme.of(context).textTheme;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Titre du stream',
          style: tt.titleSmall?.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: controller,
          enabled: enabled,
          decoration: const InputDecoration(
            hintText: 'Donne un titre à ton live...',
            prefixIcon: Icon(Icons.edit_outlined),
          ),
          maxLength: 100,
          textCapitalization: TextCapitalization.sentences,
        ),
      ],
    );
  }
}

// ── Toggle button ─────────────────────────────────────────────────────────────

class _ToggleButton extends StatelessWidget {
  final bool isStreaming;
  final VoidCallback onStart;
  final VoidCallback onStop;

  const _ToggleButton({
    required this.isStreaming,
    required this.onStart,
    required this.onStop,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    if (isStreaming) {
      return FilledButton.icon(
        onPressed: onStop,
        icon: const Icon(Icons.stop_rounded),
        label: const Text('Arrêter le stream'),
        style: FilledButton.styleFrom(
          backgroundColor: cs.error,
          foregroundColor: cs.onError,
          minimumSize: const Size.fromHeight(56),
        ),
      );
    }

    return FilledButton.icon(
      onPressed: onStart,
      icon: const Icon(Icons.fiber_manual_record_rounded),
      label: const Text('Démarrer le live'),
      style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(56)),
    );
  }
}

// ── Loading ───────────────────────────────────────────────────────────────────

class _LoadingCard extends StatelessWidget {
  const _LoadingCard();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return SizedBox(
      height: 56,
      child: Center(child: CircularProgressIndicator(color: cs.primary)),
    );
  }
}

// ── Error banner ──────────────────────────────────────────────────────────────

class _ErrorBanner extends StatelessWidget {
  final String message;
  final VoidCallback onDismiss;
  const _ErrorBanner({required this.message, required this.onDismiss});

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: cs.errorContainer,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          Icon(Icons.error_outline_rounded, color: cs.error, size: 20),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: TextStyle(color: cs.onErrorContainer, fontSize: 14),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.close_rounded, size: 18),
            onPressed: onDismiss,
            color: cs.onErrorContainer,
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
          ),
        ],
      ),
    );
  }
}

// ── Unauthorized ──────────────────────────────────────────────────────────────

class _UnauthorizedView extends StatelessWidget {
  const _UnauthorizedView();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final topPadding = MediaQuery.of(context).padding.top;

    return Scaffold(
      body: Center(
        child: Padding(
          padding: EdgeInsets.fromLTRB(40, topPadding + 40, 40, 40),
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
                child: Icon(
                  Icons.lock_outline_rounded,
                  size: 30,
                  color: cs.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 20),
              Text(
                'Accès réservé',
                style: tt.titleLarge?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text(
                'Seuls les diffuseurs peuvent accéder au studio.',
                style: tt.bodyMedium?.copyWith(color: cs.onSurfaceVariant),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
