// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'listen_history_model.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
  'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models',
);

ListenHistoryModel _$ListenHistoryModelFromJson(Map<String, dynamic> json) {
  return _ListenHistoryModel.fromJson(json);
}

/// @nodoc
mixin _$ListenHistoryModel {
  String get id => throw _privateConstructorUsedError;
  String get userId => throw _privateConstructorUsedError;
  String get streamId => throw _privateConstructorUsedError;
  String get broadcasterId => throw _privateConstructorUsedError;
  @JsonKey(name: 'timestamp')
  DateTime get timestamp => throw _privateConstructorUsedError;
  @JsonKey(name: 'duration_seconds')
  int get durationSeconds => throw _privateConstructorUsedError;

  /// Serializes this ListenHistoryModel to a JSON map.
  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;

  /// Create a copy of ListenHistoryModel
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $ListenHistoryModelCopyWith<ListenHistoryModel> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ListenHistoryModelCopyWith<$Res> {
  factory $ListenHistoryModelCopyWith(
    ListenHistoryModel value,
    $Res Function(ListenHistoryModel) then,
  ) = _$ListenHistoryModelCopyWithImpl<$Res, ListenHistoryModel>;
  @useResult
  $Res call({
    String id,
    String userId,
    String streamId,
    String broadcasterId,
    @JsonKey(name: 'timestamp') DateTime timestamp,
    @JsonKey(name: 'duration_seconds') int durationSeconds,
  });
}

/// @nodoc
class _$ListenHistoryModelCopyWithImpl<$Res, $Val extends ListenHistoryModel>
    implements $ListenHistoryModelCopyWith<$Res> {
  _$ListenHistoryModelCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of ListenHistoryModel
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? userId = null,
    Object? streamId = null,
    Object? broadcasterId = null,
    Object? timestamp = null,
    Object? durationSeconds = null,
  }) {
    return _then(
      _value.copyWith(
            id: null == id
                ? _value.id
                : id // ignore: cast_nullable_to_non_nullable
                      as String,
            userId: null == userId
                ? _value.userId
                : userId // ignore: cast_nullable_to_non_nullable
                      as String,
            streamId: null == streamId
                ? _value.streamId
                : streamId // ignore: cast_nullable_to_non_nullable
                      as String,
            broadcasterId: null == broadcasterId
                ? _value.broadcasterId
                : broadcasterId // ignore: cast_nullable_to_non_nullable
                      as String,
            timestamp: null == timestamp
                ? _value.timestamp
                : timestamp // ignore: cast_nullable_to_non_nullable
                      as DateTime,
            durationSeconds: null == durationSeconds
                ? _value.durationSeconds
                : durationSeconds // ignore: cast_nullable_to_non_nullable
                      as int,
          )
          as $Val,
    );
  }
}

/// @nodoc
abstract class _$$ListenHistoryModelImplCopyWith<$Res>
    implements $ListenHistoryModelCopyWith<$Res> {
  factory _$$ListenHistoryModelImplCopyWith(
    _$ListenHistoryModelImpl value,
    $Res Function(_$ListenHistoryModelImpl) then,
  ) = __$$ListenHistoryModelImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({
    String id,
    String userId,
    String streamId,
    String broadcasterId,
    @JsonKey(name: 'timestamp') DateTime timestamp,
    @JsonKey(name: 'duration_seconds') int durationSeconds,
  });
}

/// @nodoc
class __$$ListenHistoryModelImplCopyWithImpl<$Res>
    extends _$ListenHistoryModelCopyWithImpl<$Res, _$ListenHistoryModelImpl>
    implements _$$ListenHistoryModelImplCopyWith<$Res> {
  __$$ListenHistoryModelImplCopyWithImpl(
    _$ListenHistoryModelImpl _value,
    $Res Function(_$ListenHistoryModelImpl) _then,
  ) : super(_value, _then);

  /// Create a copy of ListenHistoryModel
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? userId = null,
    Object? streamId = null,
    Object? broadcasterId = null,
    Object? timestamp = null,
    Object? durationSeconds = null,
  }) {
    return _then(
      _$ListenHistoryModelImpl(
        id: null == id
            ? _value.id
            : id // ignore: cast_nullable_to_non_nullable
                  as String,
        userId: null == userId
            ? _value.userId
            : userId // ignore: cast_nullable_to_non_nullable
                  as String,
        streamId: null == streamId
            ? _value.streamId
            : streamId // ignore: cast_nullable_to_non_nullable
                  as String,
        broadcasterId: null == broadcasterId
            ? _value.broadcasterId
            : broadcasterId // ignore: cast_nullable_to_non_nullable
                  as String,
        timestamp: null == timestamp
            ? _value.timestamp
            : timestamp // ignore: cast_nullable_to_non_nullable
                  as DateTime,
        durationSeconds: null == durationSeconds
            ? _value.durationSeconds
            : durationSeconds // ignore: cast_nullable_to_non_nullable
                  as int,
      ),
    );
  }
}

/// @nodoc
@JsonSerializable()
class _$ListenHistoryModelImpl implements _ListenHistoryModel {
  const _$ListenHistoryModelImpl({
    required this.id,
    required this.userId,
    required this.streamId,
    required this.broadcasterId,
    @JsonKey(name: 'timestamp') required this.timestamp,
    @JsonKey(name: 'duration_seconds') this.durationSeconds = 0,
  });

  factory _$ListenHistoryModelImpl.fromJson(Map<String, dynamic> json) =>
      _$$ListenHistoryModelImplFromJson(json);

  @override
  final String id;
  @override
  final String userId;
  @override
  final String streamId;
  @override
  final String broadcasterId;
  @override
  @JsonKey(name: 'timestamp')
  final DateTime timestamp;
  @override
  @JsonKey(name: 'duration_seconds')
  final int durationSeconds;

  @override
  String toString() {
    return 'ListenHistoryModel(id: $id, userId: $userId, streamId: $streamId, broadcasterId: $broadcasterId, timestamp: $timestamp, durationSeconds: $durationSeconds)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ListenHistoryModelImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.streamId, streamId) ||
                other.streamId == streamId) &&
            (identical(other.broadcasterId, broadcasterId) ||
                other.broadcasterId == broadcasterId) &&
            (identical(other.timestamp, timestamp) ||
                other.timestamp == timestamp) &&
            (identical(other.durationSeconds, durationSeconds) ||
                other.durationSeconds == durationSeconds));
  }

  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  int get hashCode => Object.hash(
    runtimeType,
    id,
    userId,
    streamId,
    broadcasterId,
    timestamp,
    durationSeconds,
  );

  /// Create a copy of ListenHistoryModel
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$ListenHistoryModelImplCopyWith<_$ListenHistoryModelImpl> get copyWith =>
      __$$ListenHistoryModelImplCopyWithImpl<_$ListenHistoryModelImpl>(
        this,
        _$identity,
      );

  @override
  Map<String, dynamic> toJson() {
    return _$$ListenHistoryModelImplToJson(this);
  }
}

abstract class _ListenHistoryModel implements ListenHistoryModel {
  const factory _ListenHistoryModel({
    required final String id,
    required final String userId,
    required final String streamId,
    required final String broadcasterId,
    @JsonKey(name: 'timestamp') required final DateTime timestamp,
    @JsonKey(name: 'duration_seconds') final int durationSeconds,
  }) = _$ListenHistoryModelImpl;

  factory _ListenHistoryModel.fromJson(Map<String, dynamic> json) =
      _$ListenHistoryModelImpl.fromJson;

  @override
  String get id;
  @override
  String get userId;
  @override
  String get streamId;
  @override
  String get broadcasterId;
  @override
  @JsonKey(name: 'timestamp')
  DateTime get timestamp;
  @override
  @JsonKey(name: 'duration_seconds')
  int get durationSeconds;

  /// Create a copy of ListenHistoryModel
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$ListenHistoryModelImplCopyWith<_$ListenHistoryModelImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
