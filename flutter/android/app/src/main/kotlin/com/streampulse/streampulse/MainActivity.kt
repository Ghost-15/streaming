package com.streampulse.streampulse

import com.ryanheise.audioservice.AudioServiceActivity

// just_audio_background requires the host Activity to be an AudioServiceActivity
// so the background media session binds to the correct FlutterEngine.
class MainActivity : AudioServiceActivity()
