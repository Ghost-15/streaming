// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'audio_state.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
  'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models',
);

/// @nodoc
mixin _$AudioState {
  AudioPlaybackState get playbackState => throw _privateConstructorUsedError;
  double get volume => throw _privateConstructorUsedError;
  Duration get position => throw _privateConstructorUsedError;
  Duration get duration => throw _privateConstructorUsedError;
  StreamEntity? get currentStream => throw _privateConstructorUsedError;
  String get errorMessage => throw _privateConstructorUsedError;
  bool get isShuffled => throw _privateConstructorUsedError;
  bool get isLooping => throw _privateConstructorUsedError;
  List<StreamEntity> get playlist => throw _privateConstructorUsedError;
  int get playlistIndex => throw _privateConstructorUsedError;

  /// Create a copy of AudioState
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $AudioStateCopyWith<AudioState> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $AudioStateCopyWith<$Res> {
  factory $AudioStateCopyWith(
    AudioState value,
    $Res Function(AudioState) then,
  ) = _$AudioStateCopyWithImpl<$Res, AudioState>;
  @useResult
  $Res call({
    AudioPlaybackState playbackState,
    double volume,
    Duration position,
    Duration duration,
    StreamEntity? currentStream,
    String errorMessage,
    bool isShuffled,
    bool isLooping,
    List<StreamEntity> playlist,
    int playlistIndex,
  });
}

/// @nodoc
class _$AudioStateCopyWithImpl<$Res, $Val extends AudioState>
    implements $AudioStateCopyWith<$Res> {
  _$AudioStateCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of AudioState
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? playbackState = null,
    Object? volume = null,
    Object? position = null,
    Object? duration = null,
    Object? currentStream = freezed,
    Object? errorMessage = null,
    Object? isShuffled = null,
    Object? isLooping = null,
    Object? playlist = null,
    Object? playlistIndex = null,
  }) {
    return _then(
      _value.copyWith(
            playbackState: null == playbackState
                ? _value.playbackState
                : playbackState // ignore: cast_nullable_to_non_nullable
                      as AudioPlaybackState,
            volume: null == volume
                ? _value.volume
                : volume // ignore: cast_nullable_to_non_nullable
                      as double,
            position: null == position
                ? _value.position
                : position // ignore: cast_nullable_to_non_nullable
                      as Duration,
            duration: null == duration
                ? _value.duration
                : duration // ignore: cast_nullable_to_non_nullable
                      as Duration,
            currentStream: freezed == currentStream
                ? _value.currentStream
                : currentStream // ignore: cast_nullable_to_non_nullable
                      as StreamEntity?,
            errorMessage: null == errorMessage
                ? _value.errorMessage
                : errorMessage // ignore: cast_nullable_to_non_nullable
                      as String,
            isShuffled: null == isShuffled
                ? _value.isShuffled
                : isShuffled // ignore: cast_nullable_to_non_nullable
                      as bool,
            isLooping: null == isLooping
                ? _value.isLooping
                : isLooping // ignore: cast_nullable_to_non_nullable
                      as bool,
            playlist: null == playlist
                ? _value.playlist
                : playlist // ignore: cast_nullable_to_non_nullable
                      as List<StreamEntity>,
            playlistIndex: null == playlistIndex
                ? _value.playlistIndex
                : playlistIndex // ignore: cast_nullable_to_non_nullable
                      as int,
          )
          as $Val,
    );
  }
}

/// @nodoc
abstract class _$$AudioStateImplCopyWith<$Res>
    implements $AudioStateCopyWith<$Res> {
  factory _$$AudioStateImplCopyWith(
    _$AudioStateImpl value,
    $Res Function(_$AudioStateImpl) then,
  ) = __$$AudioStateImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({
    AudioPlaybackState playbackState,
    double volume,
    Duration position,
    Duration duration,
    StreamEntity? currentStream,
    String errorMessage,
    bool isShuffled,
    bool isLooping,
    List<StreamEntity> playlist,
    int playlistIndex,
  });
}

/// @nodoc
class __$$AudioStateImplCopyWithImpl<$Res>
    extends _$AudioStateCopyWithImpl<$Res, _$AudioStateImpl>
    implements _$$AudioStateImplCopyWith<$Res> {
  __$$AudioStateImplCopyWithImpl(
    _$AudioStateImpl _value,
    $Res Function(_$AudioStateImpl) _then,
  ) : super(_value, _then);

  /// Create a copy of AudioState
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? playbackState = null,
    Object? volume = null,
    Object? position = null,
    Object? duration = null,
    Object? currentStream = freezed,
    Object? errorMessage = null,
    Object? isShuffled = null,
    Object? isLooping = null,
    Object? playlist = null,
    Object? playlistIndex = null,
  }) {
    return _then(
      _$AudioStateImpl(
        playbackState: null == playbackState
            ? _value.playbackState
            : playbackState // ignore: cast_nullable_to_non_nullable
                  as AudioPlaybackState,
        volume: null == volume
            ? _value.volume
            : volume // ignore: cast_nullable_to_non_nullable
                  as double,
        position: null == position
            ? _value.position
            : position // ignore: cast_nullable_to_non_nullable
                  as Duration,
        duration: null == duration
            ? _value.duration
            : duration // ignore: cast_nullable_to_non_nullable
                  as Duration,
        currentStream: freezed == currentStream
            ? _value.currentStream
            : currentStream // ignore: cast_nullable_to_non_nullable
                  as StreamEntity?,
        errorMessage: null == errorMessage
            ? _value.errorMessage
            : errorMessage // ignore: cast_nullable_to_non_nullable
                  as String,
        isShuffled: null == isShuffled
            ? _value.isShuffled
            : isShuffled // ignore: cast_nullable_to_non_nullable
                  as bool,
        isLooping: null == isLooping
            ? _value.isLooping
            : isLooping // ignore: cast_nullable_to_non_nullable
                  as bool,
        playlist: null == playlist
            ? _value._playlist
            : playlist // ignore: cast_nullable_to_non_nullable
                  as List<StreamEntity>,
        playlistIndex: null == playlistIndex
            ? _value.playlistIndex
            : playlistIndex // ignore: cast_nullable_to_non_nullable
                  as int,
      ),
    );
  }
}

/// @nodoc

class _$AudioStateImpl extends _AudioState {
  const _$AudioStateImpl({
    required this.playbackState,
    required this.volume,
    required this.position,
    required this.duration,
    this.currentStream,
    this.errorMessage = '',
    this.isShuffled = false,
    this.isLooping = false,
    final List<StreamEntity> playlist = const [],
    this.playlistIndex = 0,
  }) : _playlist = playlist,
       super._();

  @override
  final AudioPlaybackState playbackState;
  @override
  final double volume;
  @override
  final Duration position;
  @override
  final Duration duration;
  @override
  final StreamEntity? currentStream;
  @override
  @JsonKey()
  final String errorMessage;
  @override
  @JsonKey()
  final bool isShuffled;
  @override
  @JsonKey()
  final bool isLooping;
  final List<StreamEntity> _playlist;
  @override
  @JsonKey()
  List<StreamEntity> get playlist {
    if (_playlist is EqualUnmodifiableListView) return _playlist;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_playlist);
  }

  @override
  @JsonKey()
  final int playlistIndex;

  @override
  String toString() {
    return 'AudioState(playbackState: $playbackState, volume: $volume, position: $position, duration: $duration, currentStream: $currentStream, errorMessage: $errorMessage, isShuffled: $isShuffled, isLooping: $isLooping, playlist: $playlist, playlistIndex: $playlistIndex)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$AudioStateImpl &&
            (identical(other.playbackState, playbackState) ||
                other.playbackState == playbackState) &&
            (identical(other.volume, volume) || other.volume == volume) &&
            (identical(other.position, position) ||
                other.position == position) &&
            (identical(other.duration, duration) ||
                other.duration == duration) &&
            (identical(other.currentStream, currentStream) ||
                other.currentStream == currentStream) &&
            (identical(other.errorMessage, errorMessage) ||
                other.errorMessage == errorMessage) &&
            (identical(other.isShuffled, isShuffled) ||
                other.isShuffled == isShuffled) &&
            (identical(other.isLooping, isLooping) ||
                other.isLooping == isLooping) &&
            const DeepCollectionEquality().equals(other._playlist, _playlist) &&
            (identical(other.playlistIndex, playlistIndex) ||
                other.playlistIndex == playlistIndex));
  }

  @override
  int get hashCode => Object.hash(
    runtimeType,
    playbackState,
    volume,
    position,
    duration,
    currentStream,
    errorMessage,
    isShuffled,
    isLooping,
    const DeepCollectionEquality().hash(_playlist),
    playlistIndex,
  );

  /// Create a copy of AudioState
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$AudioStateImplCopyWith<_$AudioStateImpl> get copyWith =>
      __$$AudioStateImplCopyWithImpl<_$AudioStateImpl>(this, _$identity);
}

abstract class _AudioState extends AudioState {
  const factory _AudioState({
    required final AudioPlaybackState playbackState,
    required final double volume,
    required final Duration position,
    required final Duration duration,
    final StreamEntity? currentStream,
    final String errorMessage,
    final bool isShuffled,
    final bool isLooping,
    final List<StreamEntity> playlist,
    final int playlistIndex,
  }) = _$AudioStateImpl;
  const _AudioState._() : super._();

  @override
  AudioPlaybackState get playbackState;
  @override
  double get volume;
  @override
  Duration get position;
  @override
  Duration get duration;
  @override
  StreamEntity? get currentStream;
  @override
  String get errorMessage;
  @override
  bool get isShuffled;
  @override
  bool get isLooping;
  @override
  List<StreamEntity> get playlist;
  @override
  int get playlistIndex;

  /// Create a copy of AudioState
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$AudioStateImplCopyWith<_$AudioStateImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
