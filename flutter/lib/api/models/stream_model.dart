class StreamModel {
  final String id;
  final String title;
  final String broadcasterId;
  final String broadcasterName;
  final int listenerCount;
  final String description;
  final String streamUrl;
  final bool isLive;
  final DateTime createdAt;

  const StreamModel({
    required this.id,
    required this.title,
    required this.broadcasterId,
    required this.broadcasterName,
    this.listenerCount = 0,
    this.description = '',
    required this.streamUrl,
    this.isLive = false,
    required this.createdAt,
  });

  factory StreamModel.fromJson(Map<String, dynamic> json) {
    return StreamModel(
      id: json['id'],
      title: json['title'],
      broadcasterId: json['broadcasterId'] ?? json['broadcaster_id'] ?? '',
      broadcasterName: json['broadcasterName'] ?? json['broadcaster_name'] ?? '',
      listenerCount: json['listenerCount'] ?? json['listener_count'] ?? 0,
      description: json['description'] ?? '',
      streamUrl: json['streamUrl'] ?? json['stream_url'] ?? '',
      // Backend exposes a `status` enum ("live"/"ended"); derive isLive from it,
      // with a fallback to the legacy isLive/is_live boolean.
      isLive: json['status'] != null
          ? json['status'] == 'live'
          : (json['isLive'] ?? json['is_live'] ?? false),
      createdAt: DateTime.tryParse(
            json['createdAt'] ?? json['created_at'] ?? '',
          ) ??
          DateTime.now(),
    );
  }
}
