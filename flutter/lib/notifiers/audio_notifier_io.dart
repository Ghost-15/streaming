import 'dart:async';
import 'dart:io';
import 'dart:math' as math;

import 'package:audio_session/audio_session.dart';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:just_audio/just_audio.dart';
import 'package:just_audio_background/just_audio_background.dart';

import '../api/models/stream_model.dart';
import '../api/repositories/stream_repository.dart';

enum AudioPlaybackState { idle, loading, buffering, playing, paused, error }

/// Mobile/desktop (dart:io) live audio listener backed by just_audio.
///
/// The public surface mirrors the Web implementation (`audio_notifier_web.dart`)
/// exactly so the UI and providers stay platform-agnostic. just_audio buffers
/// the live `GET /streams/:id/audio` response (natively decodable by ExoPlayer
/// on Android). A `MediaItem` tag drives the lock-screen / notification controls
/// and background playback, while `audio_session` handles OS audio-focus
/// interruptions (ducking on a transient sound, pausing on a phone call, and
/// pausing when headphones are unplugged).
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
  // handleInterruptions is disabled so this notifier fully owns audio-focus
  // behaviour, including volume ducking which just_audio does not do on its own.
  final AudioPlayer _player = AudioPlayer(handleInterruptions: false);
  StreamSubscription<PlayerState>? _stateSub;
  StreamSubscription<AudioInterruptionEvent>? _interruptionSub;
  StreamSubscription<void>? _noisySub;
  StreamSubscription<PlaybackEvent>? _eventSub;
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
    unawaited(_player.setVolume(_volume));
    _stateSub = _player.playerStateStream.listen(_onPlayerState);
    // A dropped live connection surfaces as a playback-event error rather than a
    // state change, so recovery is driven from here.
    _eventSub = _player.playbackEventStream.listen((_) {}, onError: _onPlaybackError);
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
          unawaited(_player.setVolume(_volume * _duckVolumeFactor));
        case AudioInterruptionType.pause:
        case AudioInterruptionType.unknown:
          // A phone call or another app takes exclusive focus: pause.
          _interruptionPaused = isPlaying || isBuffering || isLoading;
          unawaited(_player.pause());
          if (_interruptionPaused) _setState(AudioPlaybackState.paused);
      }
    } else {
      switch (event.type) {
        case AudioInterruptionType.duck:
          unawaited(_player.setVolume(_volume));
        case AudioInterruptionType.pause:
          if (_interruptionPaused && !_userPaused) unawaited(resume());
          _interruptionPaused = false;
        case AudioInterruptionType.unknown:
          _interruptionPaused = false;
      }
    }
  }

  void _onPlayerState(PlayerState state) {
    if (_disposed || hasError) return;
    switch (state.processingState) {
      case ProcessingState.idle:
        // Ignore transient idle bursts emitted while (re)loading a source; our
        // own stop()/error paths own the idle and error transitions.
        return;
      case ProcessingState.loading:
        _setState(AudioPlaybackState.loading);
      case ProcessingState.buffering:
        _setState(
          _userPaused
              ? AudioPlaybackState.paused
              : AudioPlaybackState.buffering,
        );
      case ProcessingState.ready:
        if (state.playing) _reconnectAttempt = 0;
        _setState(
          state.playing
              ? AudioPlaybackState.playing
              : AudioPlaybackState.paused,
        );
      case ProcessingState.completed:
        _finish();
    }
  }

  void _onPlaybackError(Object error, StackTrace stack) {
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

  // On mobile the source carries a MediaItem so the OS shows lock-screen and
  // notification controls; other io targets (desktop) play without it.
  Future<void> _loadLiveSource(StreamModel stream) {
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
      await _loadLiveSource(stream);
      await _player.play();
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
    await _player.pause();
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
      await _loadLiveSource(stream);
      await _player.play();
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
    await _player.stop();
    _leaveJoinedStream();
    _currentStream = null;
    _setState(AudioPlaybackState.idle);
  }

  Future<void> setVolume(double volume) async {
    _volume = volume.clamp(0.0, 1.0);
    await _player.setVolume(_volume);
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
    unawaited(_stateSub?.cancel());
    unawaited(_interruptionSub?.cancel());
    unawaited(_noisySub?.cancel());
    unawaited(_eventSub?.cancel());
    unawaited(_connSub?.cancel());
    unawaited(_player.dispose());
    super.dispose();
  }
}
