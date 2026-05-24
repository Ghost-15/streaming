import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../domain/entities/stream.dart';
import '../../data/repositories/stream_repository_impl.dart';

/// Provider for fetching active streams
final activeStreamsProvider = FutureProvider<List<StreamEntity>>((ref) async {
  final streamRepository = ref.watch(streamRepositoryProvider);
  return streamRepository.getActiveStreams();
});

/// Provider for stream repository
final streamRepositoryProvider = Provider<StreamRepositoryImpl>((ref) {
  return StreamRepositoryImpl();
});

/// Provider for a specific stream by ID
final streamByIdProvider = FutureProvider.family<StreamEntity, String>((ref, streamId) async {
  final streamRepository = ref.watch(streamRepositoryProvider);
  return streamRepository.getStreamById(streamId);
});

/// Provider for joined streams (future: store user's joined streams)
final joinedStreamsProvider = StateProvider<List<StreamEntity>>((ref) {
  return [];
});

/// Provider for current listening history
final listeningHistoryProvider = StateProvider<List<String>>((ref) {
  return [];
});
