import 'package:freezed_annotation/freezed_annotation.dart';

part 'stream_model.freezed.dart';
part 'stream_model.g.dart';

/// Data Transfer Object for Stream
@freezed
class StreamModel with _$StreamModel {
  const factory StreamModel({
    required String id,
    required String title,
    required String broadcasterId,
    required String broadcasterName,
    @Default(0) int listenerCount,
    @Default('') String description,
    @Default('') String thumbnail,
    @Default(false) bool isLive,
    @JsonKey(name: 'stream_url') required String streamUrl,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') required DateTime updatedAt,
  }) = _StreamModel;

  factory StreamModel.fromJson(Map<String, dynamic> json) =>
      _$StreamModelFromJson(json);
}
