import 'dart:io';

import 'package:just_audio_background/just_audio_background.dart';
import 'package:media_kit/media_kit.dart';

/// Initializes the background media session so audio keeps playing when the app
/// is backgrounded or the screen is locked, with lock-screen and notification
/// controls. Only Android and iOS ship a background media service. On iOS the
/// live WebM/Opus stream is decoded by media_kit (libmpv), which must be
/// initialized once before any player is created.
Future<void> initAudioBackground() async {
  if (!Platform.isAndroid && !Platform.isIOS) return;
  MediaKit.ensureInitialized();
  await JustAudioBackground.init(
    androidNotificationChannelId: 'fr.streampulse.audio',
    androidNotificationChannelName: 'Lecture StreamPulse',
    androidNotificationOngoing: true,
  );
}
