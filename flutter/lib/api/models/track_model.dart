class TrackModel {
  final String id;
  final String title;
  final String artist;
  final int duration;
  final String fileUrl;
  final String uploadedBy;
  final int position;
  final DateTime createdAt;

  const TrackModel({
    required this.id,
    required this.title,
    this.artist = '',
    this.duration = 0,
    this.fileUrl = '',
    this.uploadedBy = '',
    this.position = 0,
    required this.createdAt,
  });

  factory TrackModel.fromJson(Map<String, dynamic> json) {
    return TrackModel(
      id: json['id'] ?? json['track_id'] ?? '',
      title: json['title'] ?? 'Titre sans nom',
      artist: json['artist'] ?? '',
      duration: json['duration'] ?? 0,
      fileUrl: json['fileUrl'] ?? json['file_url'] ?? '',
      uploadedBy: json['uploadedBy'] ?? json['uploaded_by'] ?? '',
      position: json['position'] ?? 0,
      createdAt:
          DateTime.tryParse(
            json['createdAt'] ?? json['created_at'] ?? json['added_at'] ?? '',
          ) ??
          DateTime.now(),
    );
  }
}
