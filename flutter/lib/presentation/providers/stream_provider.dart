import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/stream.dart';

// ── Active streams list ───────────────────────────────────────────────────────

class ActiveStreamsNotifier extends Notifier<List<StreamEntity>> {
  @override
  List<StreamEntity> build() => [];
}

final activeStreamsProvider =
    NotifierProvider<ActiveStreamsNotifier, List<StreamEntity>>(
        ActiveStreamsNotifier.new);

// ── Currently joined stream ───────────────────────────────────────────────────

class ActiveStreamNotifier extends Notifier<StreamEntity?> {
  @override
  StreamEntity? build() => null;
}

final activeStreamProvider =
    NotifierProvider<ActiveStreamNotifier, StreamEntity?>(
        ActiveStreamNotifier.new);
