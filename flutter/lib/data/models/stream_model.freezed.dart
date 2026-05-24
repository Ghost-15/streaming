// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'stream_model.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
  'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models',
);

StreamModel _$StreamModelFromJson(Map<String, dynamic> json) {
  return _StreamModel.fromJson(json);
}

/// @nodoc
mixin _$StreamModel {
  String get id => throw _privateConstructorUsedError;
  String get title => throw _privateConstructorUsedError;
  String get broadcasterId => throw _privateConstructorUsedError;
  String get broadcasterName => throw _privateConstructorUsedError;
  int get listenerCount => throw _privateConstructorUsedError;
  String get description => throw _privateConstructorUsedError;
  String get thumbnail => throw _privateConstructorUsedError;
  bool get isLive => throw _privateConstructorUsedError;
  @JsonKey(name: 'stream_url')
  String get streamUrl => throw _privateConstructorUsedError;
  @JsonKey(name: 'created_at')
  DateTime get createdAt => throw _privateConstructorUsedError;
  @JsonKey(name: 'updated_at')
  DateTime get updatedAt => throw _privateConstructorUsedError;

  /// Serializes this StreamModel to a JSON map.
  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;

  /// Create a copy of StreamModel
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $StreamModelCopyWith<StreamModel> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $StreamModelCopyWith<$Res> {
  factory $StreamModelCopyWith(
    StreamModel value,
    $Res Function(StreamModel) then,
  ) = _$StreamModelCopyWithImpl<$Res, StreamModel>;
  @useResult
  $Res call({
    String id,
    String title,
    String broadcasterId,
    String broadcasterName,
    int listenerCount,
    String description,
    String thumbnail,
    bool isLive,
    @JsonKey(name: 'stream_url') String streamUrl,
    @JsonKey(name: 'created_at') DateTime createdAt,
    @JsonKey(name: 'updated_at') DateTime updatedAt,
  });
}

/// @nodoc
class _$StreamModelCopyWithImpl<$Res, $Val extends StreamModel>
    implements $StreamModelCopyWith<$Res> {
  _$StreamModelCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of StreamModel
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? title = null,
    Object? broadcasterId = null,
    Object? broadcasterName = null,
    Object? listenerCount = null,
    Object? description = null,
    Object? thumbnail = null,
    Object? isLive = null,
    Object? streamUrl = null,
    Object? createdAt = null,
    Object? updatedAt = null,
  }) {
    return _then(
      _value.copyWith(
            id: null == id
                ? _value.id
                : id // ignore: cast_nullable_to_non_nullable
                      as String,
            title: null == title
                ? _value.title
                : title // ignore: cast_nullable_to_non_nullable
                      as String,
            broadcasterId: null == broadcasterId
                ? _value.broadcasterId
                : broadcasterId // ignore: cast_nullable_to_non_nullable
                      as String,
            broadcasterName: null == broadcasterName
                ? _value.broadcasterName
                : broadcasterName // ignore: cast_nullable_to_non_nullable
                      as String,
            listenerCount: null == listenerCount
                ? _value.listenerCount
                : listenerCount // ignore: cast_nullable_to_non_nullable
                      as int,
            description: null == description
                ? _value.description
                : description // ignore: cast_nullable_to_non_nullable
                      as String,
            thumbnail: null == thumbnail
                ? _value.thumbnail
                : thumbnail // ignore: cast_nullable_to_non_nullable
                      as String,
            isLive: null == isLive
                ? _value.isLive
                : isLive // ignore: cast_nullable_to_non_nullable
                      as bool,
            streamUrl: null == streamUrl
                ? _value.streamUrl
                : streamUrl // ignore: cast_nullable_to_non_nullable
                      as String,
            createdAt: null == createdAt
                ? _value.createdAt
                : createdAt // ignore: cast_nullable_to_non_nullable
                      as DateTime,
            updatedAt: null == updatedAt
                ? _value.updatedAt
                : updatedAt // ignore: cast_nullable_to_non_nullable
                      as DateTime,
          )
          as $Val,
    );
  }
}

/// @nodoc
abstract class _$$StreamModelImplCopyWith<$Res>
    implements $StreamModelCopyWith<$Res> {
  factory _$$StreamModelImplCopyWith(
    _$StreamModelImpl value,
    $Res Function(_$StreamModelImpl) then,
  ) = __$$StreamModelImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({
    String id,
    String title,
    String broadcasterId,
    String broadcasterName,
    int listenerCount,
    String description,
    String thumbnail,
    bool isLive,
    @JsonKey(name: 'stream_url') String streamUrl,
    @JsonKey(name: 'created_at') DateTime createdAt,
    @JsonKey(name: 'updated_at') DateTime updatedAt,
  });
}

/// @nodoc
class __$$StreamModelImplCopyWithImpl<$Res>
    extends _$StreamModelCopyWithImpl<$Res, _$StreamModelImpl>
    implements _$$StreamModelImplCopyWith<$Res> {
  __$$StreamModelImplCopyWithImpl(
    _$StreamModelImpl _value,
    $Res Function(_$StreamModelImpl) _then,
  ) : super(_value, _then);

  /// Create a copy of StreamModel
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? title = null,
    Object? broadcasterId = null,
    Object? broadcasterName = null,
    Object? listenerCount = null,
    Object? description = null,
    Object? thumbnail = null,
    Object? isLive = null,
    Object? streamUrl = null,
    Object? createdAt = null,
    Object? updatedAt = null,
  }) {
    return _then(
      _$StreamModelImpl(
        id: null == id
            ? _value.id
            : id // ignore: cast_nullable_to_non_nullable
                  as String,
        title: null == title
            ? _value.title
            : title // ignore: cast_nullable_to_non_nullable
                  as String,
        broadcasterId: null == broadcasterId
            ? _value.broadcasterId
            : broadcasterId // ignore: cast_nullable_to_non_nullable
                  as String,
        broadcasterName: null == broadcasterName
            ? _value.broadcasterName
            : broadcasterName // ignore: cast_nullable_to_non_nullable
                  as String,
        listenerCount: null == listenerCount
            ? _value.listenerCount
            : listenerCount // ignore: cast_nullable_to_non_nullable
                  as int,
        description: null == description
            ? _value.description
            : description // ignore: cast_nullable_to_non_nullable
                  as String,
        thumbnail: null == thumbnail
            ? _value.thumbnail
            : thumbnail // ignore: cast_nullable_to_non_nullable
                  as String,
        isLive: null == isLive
            ? _value.isLive
            : isLive // ignore: cast_nullable_to_non_nullable
                  as bool,
        streamUrl: null == streamUrl
            ? _value.streamUrl
            : streamUrl // ignore: cast_nullable_to_non_nullable
                  as String,
        createdAt: null == createdAt
            ? _value.createdAt
            : createdAt // ignore: cast_nullable_to_non_nullable
                  as DateTime,
        updatedAt: null == updatedAt
            ? _value.updatedAt
            : updatedAt // ignore: cast_nullable_to_non_nullable
                  as DateTime,
      ),
    );
  }
}

/// @nodoc
@JsonSerializable()
class _$StreamModelImpl implements _StreamModel {
  const _$StreamModelImpl({
    required this.id,
    required this.title,
    required this.broadcasterId,
    required this.broadcasterName,
    this.listenerCount = 0,
    this.description = '',
    this.thumbnail = '',
    this.isLive = false,
    @JsonKey(name: 'stream_url') required this.streamUrl,
    @JsonKey(name: 'created_at') required this.createdAt,
    @JsonKey(name: 'updated_at') required this.updatedAt,
  });

  factory _$StreamModelImpl.fromJson(Map<String, dynamic> json) =>
      _$$StreamModelImplFromJson(json);

  @override
  final String id;
  @override
  final String title;
  @override
  final String broadcasterId;
  @override
  final String broadcasterName;
  @override
  @JsonKey()
  final int listenerCount;
  @override
  @JsonKey()
  final String description;
  @override
  @JsonKey()
  final String thumbnail;
  @override
  @JsonKey()
  final bool isLive;
  @override
  @JsonKey(name: 'stream_url')
  final String streamUrl;
  @override
  @JsonKey(name: 'created_at')
  final DateTime createdAt;
  @override
  @JsonKey(name: 'updated_at')
  final DateTime updatedAt;

  @override
  String toString() {
    return 'StreamModel(id: $id, title: $title, broadcasterId: $broadcasterId, broadcasterName: $broadcasterName, listenerCount: $listenerCount, description: $description, thumbnail: $thumbnail, isLive: $isLive, streamUrl: $streamUrl, createdAt: $createdAt, updatedAt: $updatedAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$StreamModelImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.title, title) || other.title == title) &&
            (identical(other.broadcasterId, broadcasterId) ||
                other.broadcasterId == broadcasterId) &&
            (identical(other.broadcasterName, broadcasterName) ||
                other.broadcasterName == broadcasterName) &&
            (identical(other.listenerCount, listenerCount) ||
                other.listenerCount == listenerCount) &&
            (identical(other.description, description) ||
                other.description == description) &&
            (identical(other.thumbnail, thumbnail) ||
                other.thumbnail == thumbnail) &&
            (identical(other.isLive, isLive) || other.isLive == isLive) &&
            (identical(other.streamUrl, streamUrl) ||
                other.streamUrl == streamUrl) &&
            (identical(other.createdAt, createdAt) ||
                other.createdAt == createdAt) &&
            (identical(other.updatedAt, updatedAt) ||
                other.updatedAt == updatedAt));
  }

  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  int get hashCode => Object.hash(
    runtimeType,
    id,
    title,
    broadcasterId,
    broadcasterName,
    listenerCount,
    description,
    thumbnail,
    isLive,
    streamUrl,
    createdAt,
    updatedAt,
  );

  /// Create a copy of StreamModel
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$StreamModelImplCopyWith<_$StreamModelImpl> get copyWith =>
      __$$StreamModelImplCopyWithImpl<_$StreamModelImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$StreamModelImplToJson(this);
  }
}

abstract class _StreamModel implements StreamModel {
  const factory _StreamModel({
    required final String id,
    required final String title,
    required final String broadcasterId,
    required final String broadcasterName,
    final int listenerCount,
    final String description,
    final String thumbnail,
    final bool isLive,
    @JsonKey(name: 'stream_url') required final String streamUrl,
    @JsonKey(name: 'created_at') required final DateTime createdAt,
    @JsonKey(name: 'updated_at') required final DateTime updatedAt,
  }) = _$StreamModelImpl;

  factory _StreamModel.fromJson(Map<String, dynamic> json) =
      _$StreamModelImpl.fromJson;

  @override
  String get id;
  @override
  String get title;
  @override
  String get broadcasterId;
  @override
  String get broadcasterName;
  @override
  int get listenerCount;
  @override
  String get description;
  @override
  String get thumbnail;
  @override
  bool get isLive;
  @override
  @JsonKey(name: 'stream_url')
  String get streamUrl;
  @override
  @JsonKey(name: 'created_at')
  DateTime get createdAt;
  @override
  @JsonKey(name: 'updated_at')
  DateTime get updatedAt;

  /// Create a copy of StreamModel
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$StreamModelImplCopyWith<_$StreamModelImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
