class StreamModel {
  final String id;
  final String title;
  final String broadcasterId;
  final String broadcasterName;
  final int listenerCount;
  final String description;
  final String streamUrl;
  final String activeSessionId;
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
    this.activeSessionId = '',
    this.isLive = false,
    required this.createdAt,
  });

  factory StreamModel.fromJson(Map<String, dynamic> json) {
    return StreamModel(
      id: json['id'],
      title: json['title'],
      broadcasterId: json['broadcasterId'] ?? json['broadcaster_id'] ?? '',
      broadcasterName:
          json['broadcasterName'] ?? json['broadcaster_name'] ?? '',
      listenerCount: json['listenerCount'] ?? json['listener_count'] ?? 0,
      description: json['description'] ?? '',
      streamUrl: json['streamUrl'] ?? json['stream_url'] ?? '',
      activeSessionId:
          json['activeSessionId'] ?? json['active_session_id'] ?? '',
      isLive: json['isLive'] ?? json['is_live'] ?? json['status'] == 'live',
      createdAt:
          DateTime.tryParse(
            json['createdAt'] ?? json['created_at'] ?? json['started_at'] ?? '',
          ) ??
          DateTime.now(),
    );
  }

  StreamModel copyWith({
    String? streamUrl,
    String? activeSessionId,
    bool? isLive,
    DateTime? createdAt,
  }) {
    return StreamModel(
      id: id,
      title: title,
      broadcasterId: broadcasterId,
      broadcasterName: broadcasterName,
      listenerCount: listenerCount,
      description: description,
      streamUrl: streamUrl ?? this.streamUrl,
      activeSessionId: activeSessionId ?? this.activeSessionId,
      isLive: isLive ?? this.isLive,
      createdAt: createdAt ?? this.createdAt,
    );
  }
}
