import 'dart:async';
import 'dart:io';
import 'dart:math' as math;

import 'package:audio_session/audio_session.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:just_audio/just_audio.dart';
import 'package:just_audio_background/just_audio_background.dart';
import 'package:media_kit/media_kit.dart' as mk;

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

/// Platform-neutral status reported by a playback engine.
enum _EngineStatus { idle, loading, buffering, playing, paused, completed }

/// A minimal live-audio playback backend. Two implementations exist because the
/// live stream is WebM/Opus: ExoPlayer (Android, via just_audio) decodes it and
/// just_audio_background gives lock-screen controls, whereas AVPlayer (iOS)
/// cannot decode WebM — libmpv (via media_kit) is used there instead. The
/// surrounding notifier logic (join/leave, interruptions, reconnection) is shared.
abstract class _LiveAudioEngine {
  Stream<_EngineStatus> get statusStream;
  Stream<Object> get errorStream;
  Future<void> load(StreamModel stream);
  Future<void> play();
  Future<void> pause();
  Future<void> stop();
  Future<void> setVolume(double volume01);
  Future<void> dispose();
}

/// Android / desktop engine backed by just_audio (+ background media session).
class _JustAudioEngine implements _LiveAudioEngine {
  // handleInterruptions is disabled so the notifier fully owns audio focus,
  // including volume ducking which just_audio does not do on its own.
  final AudioPlayer _player = AudioPlayer(handleInterruptions: false);
  final StreamController<Object> _errors = StreamController<Object>.broadcast();
  StreamSubscription<PlaybackEvent>? _eventSub;

  _JustAudioEngine() {
    _eventSub = _player.playbackEventStream.listen(
      (_) {},
      onError: (Object e, StackTrace _) => _errors.add(e),
    );
  }

  @override
  Stream<_EngineStatus> get statusStream => _player.playerStateStream.map((s) {
    switch (s.processingState) {
      case ProcessingState.idle:
        return _EngineStatus.idle;
      case ProcessingState.loading:
        return _EngineStatus.loading;
      case ProcessingState.buffering:
        return _EngineStatus.buffering;
      case ProcessingState.ready:
        return s.playing ? _EngineStatus.playing : _EngineStatus.paused;
      case ProcessingState.completed:
        return _EngineStatus.completed;
    }
  });

  @override
  Stream<Object> get errorStream => _errors.stream;

  @override
  Future<void> load(StreamModel stream) {
    final uri = Uri.parse(stream.streamUrl);
    if (Platform.isAndroid || Platform.isIOS) {
      return _player.setAudioSource(
        AudioSource.uri(
          uri,
          tag: MediaItem(
            id: stream.id,
            title: stream.title,
            artist: stream.broadcasterName.isEmpty
                ? 'StreamPulse'
                : stream.broadcasterName,
            album: 'Direct StreamPulse',
          ),
        ),
      );
    }
    return _player.setAudioSource(AudioSource.uri(uri));
  }

  @override
  Future<void> play() => _player.play();
  @override
  Future<void> pause() => _player.pause();
  @override
  Future<void> stop() => _player.stop();
  @override
  Future<void> setVolume(double volume01) => _player.setVolume(volume01);

  @override
  Future<void> dispose() async {
    await _eventSub?.cancel();
    await _errors.close();
    await _player.dispose();
  }
}

/// iOS engine backed by media_kit (libmpv), which decodes the WebM/Opus live
/// stream that AVPlayer cannot. Background playback relies on the iOS `audio`
/// background mode plus the audio session configured by the notifier.
class _MediaKitEngine implements _LiveAudioEngine {
  final mk.Player _player = mk.Player();
  final StreamController<_EngineStatus> _status =
      StreamController<_EngineStatus>.broadcast();
  final List<StreamSubscription<dynamic>> _subs = [];

  _MediaKitEngine() {
    _subs.add(_player.stream.playing.listen((_) => _emit()));
    _subs.add(_player.stream.buffering.listen((_) => _emit()));
    _subs.add(_player.stream.completed.listen((_) => _emit()));
  }

  void _emit() {
    if (_status.isClosed) return;
    if (_player.state.completed) {
      _status.add(_EngineStatus.completed);
    } else if (_player.state.buffering) {
      _status.add(_EngineStatus.buffering);
    } else if (_player.state.playing) {
      _status.add(_EngineStatus.playing);
    } else {
      _status.add(_EngineStatus.paused);
    }
  }

  @override
  Stream<_EngineStatus> get statusStream => _status.stream;
  @override
  Stream<Object> get errorStream => _player.stream.error.cast<Object>();

  @override
  Future<void> load(StreamModel stream) =>
      _player.open(mk.Media(stream.streamUrl), play: false);
  @override
  Future<void> play() => _player.play();
  @override
  Future<void> pause() => _player.pause();
  @override
  Future<void> stop() => _player.stop();
  @override
  // media_kit volume is a 0-100 percentage.
  Future<void> setVolume(double volume01) => _player.setVolume(volume01 * 100);

  @override
  Future<void> dispose() async {
    for (final s in _subs) {
      await s.cancel();
    }
    await _status.close();
    await _player.dispose();
  }
}

/// Mobile/desktop (dart:io) live audio listener.
///
/// The public surface mirrors the Web implementation (`audio_notifier_web.dart`)
/// exactly so the UI and providers stay platform-agnostic. The actual decoding
/// is delegated to a [_LiveAudioEngine] chosen per platform (just_audio on
/// Android/desktop, media_kit on iOS). `audio_session` handles OS audio-focus
/// interruptions (ducking on a transient sound, pausing on a phone call, and
/// pausing when headphones are unplugged); connectivity changes trigger an
/// automatic live-edge reconnection.
class AudioNotifier extends ChangeNotifier {
  static const _duckVolumeFactor = 0.3;
  static const _maxReconnectAttempts = 8;
  static const _reconnectBackoff = <Duration>[
    Duration(milliseconds: 500),
    Duration(seconds: 1),
    Duration(seconds: 2),
    Duration(seconds: 4),
    Duration(seconds: 8),
  ];

  final StreamRepository _repository;
  final _LiveAudioEngine _engine = Platform.isIOS
      ? _MediaKitEngine()
      : _JustAudioEngine();
  StreamSubscription<_EngineStatus>? _statusSub;
  StreamSubscription<Object>? _errorSub;
  StreamSubscription<AudioInterruptionEvent>? _interruptionSub;
  StreamSubscription<void>? _noisySub;
  StreamSubscription<List<ConnectivityResult>>? _connSub;
  Timer? _reconnectTimer;
  int _reconnectAttempt = 0;

  AudioPlaybackState _playbackState = AudioPlaybackState.idle;
  double _volume = 1.0;
  StreamModel? _currentStream;
  String _errorMessage = '';
  String? _joinedStreamId;
  bool _userPaused = false;
  bool _interruptionPaused = false;
  bool _disposed = false;

  AudioNotifier({StreamRepository? repository})
    : _repository = repository ?? const StreamRepository() {
    unawaited(_engine.setVolume(_volume));
    _statusSub = _engine.statusStream.listen(_onEngineStatus);
    // A dropped live connection surfaces as an engine error, not a status change.
    _errorSub = _engine.errorStream.listen(_onPlaybackError);
    _connSub = Connectivity().onConnectivityChanged.listen(_onConnectivityChanged);
    unawaited(_configureAudioSession());
  }

  AudioPlaybackState get playbackState => _playbackState;
  double get volume => _volume;
  StreamModel? get currentStream => _currentStream;
  String get errorMessage => _errorMessage;

  bool get isPlaying => _playbackState == AudioPlaybackState.playing;
  bool get isPaused => _playbackState == AudioPlaybackState.paused;
  bool get isLoading => _playbackState == AudioPlaybackState.loading;
  bool get isBuffering => _playbackState == AudioPlaybackState.buffering;
  bool get hasError => _playbackState == AudioPlaybackState.error;

  Future<void> _configureAudioSession() async {
    try {
      final session = await AudioSession.instance;
      await session.configure(const AudioSessionConfiguration.music());
      _interruptionSub = session.interruptionEventStream.listen(_onInterruption);
      // Headphones unplugged / Bluetooth disconnected: pause like a music app.
      _noisySub = session.becomingNoisyEventStream.listen((_) => pause());
    } catch (_) {
      // Audio focus management is best-effort; playback still works without it.
    }
  }

  void _onInterruption(AudioInterruptionEvent event) {
    if (_disposed) return;
    if (event.begin) {
      switch (event.type) {
        case AudioInterruptionType.duck:
          // A short system sound (notification, GPS): lower the volume.
          unawaited(_engine.setVolume(_volume * _duckVolumeFactor));
        case AudioInterruptionType.pause:
        case AudioInterruptionType.unknown:
          // A phone call or another app takes exclusive focus: pause.
          _interruptionPaused = isPlaying || isBuffering || isLoading;
          unawaited(_engine.pause());
          if (_interruptionPaused) _setState(AudioPlaybackState.paused);
      }
    } else {
      switch (event.type) {
        case AudioInterruptionType.duck:
          unawaited(_engine.setVolume(_volume));
        case AudioInterruptionType.pause:
          if (_interruptionPaused && !_userPaused) unawaited(resume());
          _interruptionPaused = false;
        case AudioInterruptionType.unknown:
          _interruptionPaused = false;
      }
    }
  }

  void _onEngineStatus(_EngineStatus status) {
    if (_disposed || hasError) return;
    switch (status) {
      case _EngineStatus.idle:
        // Ignore transient idle bursts emitted while (re)loading a source; our
        // own stop()/error paths own the idle and error transitions.
        return;
      case _EngineStatus.loading:
        _setState(AudioPlaybackState.loading);
      case _EngineStatus.buffering:
        _setState(
          _userPaused
              ? AudioPlaybackState.paused
              : AudioPlaybackState.buffering,
        );
      case _EngineStatus.playing:
        _reconnectAttempt = 0;
        _setState(AudioPlaybackState.playing);
      case _EngineStatus.paused:
        _setState(
          _userPaused ? AudioPlaybackState.paused : AudioPlaybackState.buffering,
        );
      case _EngineStatus.completed:
        _finish();
    }
  }

  void _onPlaybackError(Object error) {
    if (_disposed || _userPaused || _currentStream == null) return;
    debugPrint('[Listener] playback error: $error');
    _scheduleReconnect();
  }

  void _onConnectivityChanged(List<ConnectivityResult> results) {
    if (_disposed) return;
    final online = results.any((r) => r != ConnectivityResult.none);
    final stream = _currentStream;
    if (stream == null || _userPaused) return;
    if (!online) {
      _setState(AudioPlaybackState.buffering);
      return;
    }
    // Network came back (e.g. Wi-Fi ↔ mobile): if playback is degraded, retry
    // right away from a fresh live edge.
    if (hasError || isBuffering) {
      _reconnectAttempt = 0;
      _scheduleReconnect(immediate: true);
    }
  }

  void _scheduleReconnect({bool immediate = false}) {
    final stream = _currentStream;
    if (stream == null || _userPaused || _disposed) return;
    if (_reconnectTimer?.isActive ?? false) return;
    if (_reconnectAttempt >= _maxReconnectAttempts) {
      _fail('Connexion au direct perdue');
      return;
    }
    _setState(AudioPlaybackState.buffering);
    final delay = immediate
        ? Duration.zero
        : _reconnectBackoff[math.min(
            _reconnectAttempt,
            _reconnectBackoff.length - 1,
          )];
    _reconnectAttempt++;
    _reconnectTimer = Timer(delay, () {
      _reconnectTimer = null;
      final current = _currentStream;
      if (current == null || _userPaused || _disposed) return;
      unawaited(playStream(current));
    });
  }

  Future<void> playStream(StreamModel stream) async {
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _userPaused = false;
    _interruptionPaused = false;
    _currentStream = stream;
    _errorMessage = '';
    _setState(AudioPlaybackState.loading);

    unawaited(_joinStream(stream.id));
    try {
      await _engine.load(stream);
      await _engine.play();
    } catch (e) {
      _fail('Impossible de lire le direct: $e');
    }
  }

  Future<void> _joinStream(String streamId) async {
    try {
      await _repository.joinStream(streamId);
      _joinedStreamId = streamId;
    } catch (_) {
      // The public audio endpoint stays usable for anonymous listeners; the
      // history/counting request is deliberately best-effort.
    }
  }

  void _leaveJoinedStream() {
    final streamId = _joinedStreamId;
    _joinedStreamId = null;
    if (streamId != null) {
      unawaited(_repository.leaveStream(streamId).catchError((_) {}));
    }
  }

  Future<void> pause() async {
    _userPaused = true;
    await _engine.pause();
    _setState(AudioPlaybackState.paused);
  }

  Future<void> resume() async {
    _userPaused = false;
    final stream = _currentStream;
    if (stream == null) return;
    // A live stream must not resume from a stale buffered position: reload so
    // playback restarts at the current live edge.
    _setState(AudioPlaybackState.loading);
    try {
      await _engine.load(stream);
      await _engine.play();
    } catch (e) {
      _fail('Lecture impossible: $e');
    }
  }

  Future<void> stop() async {
    _userPaused = false;
    _interruptionPaused = false;
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempt = 0;
    await _engine.stop();
    _leaveJoinedStream();
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0.0, 1.0);
    await _engine.setVolume(_volume);
    notifyListeners();
  }

  Future<void> retry() async {
    final stream = _currentStream;
    if (stream != null) await playStream(stream);
  }

  void clearError() {
    _errorMessage = '';
    _setState(AudioPlaybackState.idle);
  }

  void _finish() {
    _leaveJoinedStream();
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  void _fail(String message) {
    _errorMessage = message;
    _leaveJoinedStream();
    _setState(AudioPlaybackState.error);
  }

  void _setState(AudioPlaybackState state) {
    if (_disposed) return;
    _playbackState = state;
    notifyListeners();
  }

  @override
  void dispose() {
    _disposed = true;
    _reconnectTimer?.cancel();
    _leaveJoinedStream();
    unawaited(_statusSub?.cancel());
    unawaited(_errorSub?.cancel());
    unawaited(_interruptionSub?.cancel());
    unawaited(_noisySub?.cancel());
    unawaited(_connSub?.cancel());
    unawaited(_engine.dispose());
    super.dispose();
  }
}
