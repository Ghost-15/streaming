// Mobile — login screen for all non-admin users.
// Sprint 1 — US-001.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class LoginPage extends ConsumerWidget {
  const LoginPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      appBar: AppBar(title: const Text('StreamPulse')),
      body: const Center(
        // TODO Sprint 1 — US-001: email + password form, login button
        child: Text('Login — TODO Sprint 1'),
      ),
    );
  }
}
