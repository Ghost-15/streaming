import 'package:freezed_annotation/freezed_annotation.dart';

part 'listen_history_model.freezed.dart';
part 'listen_history_model.g.dart';

/// Data Transfer Object for Listen History Entry
@freezed
class ListenHistoryModel with _$ListenHistoryModel {
  const factory ListenHistoryModel({
    required String id,
    required String userId,
    required String streamId,
    required String broadcasterId,
    @JsonKey(name: 'timestamp') required DateTime timestamp,
    @JsonKey(name: 'duration_seconds') @Default(0) int durationSeconds,
  }) = _ListenHistoryModel;

  factory ListenHistoryModel.fromJson(Map<String, dynamic> json) =>
      _$ListenHistoryModelFromJson(json);
}
