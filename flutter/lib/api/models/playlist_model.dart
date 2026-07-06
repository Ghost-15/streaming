import 'track_model.dart';

class PlaylistModel {
  final String id;
  final String ownerId;
  final String title;
  final bool isQueue;
  final int trackCount;
  final DateTime createdAt;
  final List<TrackModel> tracks;

  const PlaylistModel({
    required this.id,
    required this.ownerId,
    required this.title,
    this.isQueue = false,
    this.trackCount = 0,
    required this.createdAt,
    this.tracks = const [],
  });

  factory PlaylistModel.fromJson(Map<String, dynamic> json) {
    final rawTracks = json['tracks'] as List? ?? const [];
    return PlaylistModel(
      id: json['id'] ?? '',
      ownerId: json['ownerId'] ?? json['owner_id'] ?? '',
      title: json['title'] ?? 'Playlist',
      isQueue: json['isQueue'] ?? json['is_queue'] ?? false,
      trackCount: json['trackCount'] ?? json['track_count'] ?? rawTracks.length,
      createdAt:
          DateTime.tryParse(json['createdAt'] ?? json['created_at'] ?? '') ??
          DateTime.now(),
      tracks: rawTracks
          .map((e) => TrackModel.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
