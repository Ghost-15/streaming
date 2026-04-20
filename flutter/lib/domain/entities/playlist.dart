class Playlist {
  final String id;
  final String ownerId;
  final String title;
  final bool isQueue;
  final int trackCount; // maintenu par trigger BDD (migration 005)
  final List<Track> tracks;

  const Playlist({
    required this.id,
    required this.ownerId,
    required this.title,
    required this.isQueue,
    required this.trackCount,
    this.tracks = const [],
  });
}

/// Track = fichier audio stocké dans Supabase Storage (bucket: audio).
class Track {
  final String id;
  final String title;
  final String artist;
  final int duration; // secondes
  final String fileUrl; // URL Supabase Storage
  final String uploadedBy;

  // Champs de liaison playlist (présents quand la track est dans une playlist)
  final String? playlistId;
  final int? position;

  const Track({
    required this.id,
    required this.title,
    required this.artist,
    required this.duration,
    required this.fileUrl,
    required this.uploadedBy,
    this.playlistId,
    this.position,
  });
}
