// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'listen_history_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$ListenHistoryModelImpl _$$ListenHistoryModelImplFromJson(
  Map<String, dynamic> json,
) => _$ListenHistoryModelImpl(
  id: json['id'] as String,
  userId: json['userId'] as String,
  streamId: json['streamId'] as String,
  broadcasterId: json['broadcasterId'] as String,
  timestamp: DateTime.parse(json['timestamp'] as String),
  durationSeconds: (json['duration_seconds'] as num?)?.toInt() ?? 0,
);

Map<String, dynamic> _$$ListenHistoryModelImplToJson(
  _$ListenHistoryModelImpl instance,
) => <String, dynamic>{
  'id': instance.id,
  'userId': instance.userId,
  'streamId': instance.streamId,
  'broadcasterId': instance.broadcasterId,
  'timestamp': instance.timestamp.toIso8601String(),
  'duration_seconds': instance.durationSeconds,
};
