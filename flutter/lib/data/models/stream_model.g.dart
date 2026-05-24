// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'stream_model.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$StreamModelImpl _$$StreamModelImplFromJson(Map<String, dynamic> json) =>
    _$StreamModelImpl(
      id: json['id'] as String,
      title: json['title'] as String,
      broadcasterId: json['broadcasterId'] as String,
      broadcasterName: json['broadcasterName'] as String,
      listenerCount: (json['listenerCount'] as num?)?.toInt() ?? 0,
      description: json['description'] as String? ?? '',
      thumbnail: json['thumbnail'] as String? ?? '',
      isLive: json['isLive'] as bool? ?? false,
      streamUrl: json['stream_url'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );

Map<String, dynamic> _$$StreamModelImplToJson(_$StreamModelImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'title': instance.title,
      'broadcasterId': instance.broadcasterId,
      'broadcasterName': instance.broadcasterName,
      'listenerCount': instance.listenerCount,
      'description': instance.description,
      'thumbnail': instance.thumbnail,
      'isLive': instance.isLive,
      'stream_url': instance.streamUrl,
      'created_at': instance.createdAt.toIso8601String(),
      'updated_at': instance.updatedAt.toIso8601String(),
    };
