import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app.dart';

void main() {
  runApp(
    // ProviderScope is required at the root for Riverpod.
    const ProviderScope(
      child: StreamPulseApp(),
    ),
  );
}
