import 'dart:io';

import 'package:just_audio_background/just_audio_background.dart';

/// Initializes the background media session so audio keeps playing when the app
/// is backgrounded or the screen is locked, with lock-screen and notification
/// controls. Only Android and iOS ship a background media service.
Future<void> initAudioBackground() async {
  if (!Platform.isAndroid && !Platform.isIOS) return;
  await JustAudioBackground.init(
    androidNotificationChannelId: 'fr.streampulse.audio',
    androidNotificationChannelName: 'Lecture StreamPulse',
    androidNotificationOngoing: true,
  );
}
