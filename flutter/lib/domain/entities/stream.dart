// Domain entity — zero Flutter/external dependencies.
enum StreamStatus { live, ended }

class StreamEntity {
  final String id;
  final String title;
  final String broadcasterId;
  final StreamStatus status;
  final int listenerCount;

  const StreamEntity({
    required this.id,
    required this.title,
    required this.broadcasterId,
    required this.status,
    required this.listenerCount,
  });

  bool get isLive => status == StreamStatus.live;
}
