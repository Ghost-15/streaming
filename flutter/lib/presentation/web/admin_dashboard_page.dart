// Web — admin-only dashboard. Only accessible when role == admin on Flutter Web.
// Sprint 3 — US-013.
import 'package:flutter/material.dart';

class AdminDashboardPage extends StatelessWidget {
  const AdminDashboardPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('StreamPulse Admin')),
      body: const Center(
        // TODO Sprint 3 — US-013: user management table, stream oversight
        child: Text('Admin Dashboard — TODO Sprint 3'),
      ),
    );
  }
}
