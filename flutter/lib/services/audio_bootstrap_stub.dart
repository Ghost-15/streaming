/// No-op background-audio initialization for platforms without a native
/// background media service (Web). Playback there uses the browser Media
/// Session API instead.
Future<void> initAudioBackground() async {}
