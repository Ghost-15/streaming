import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../api/models/role.dart';
import '../notifiers/session_notifier.dart';
import 'mini_player.dart';

class ScaffoldWithNavBar extends StatelessWidget {
  final StatefulNavigationShell navigationShell;
  const ScaffoldWithNavBar({super.key, required this.navigationShell});

  @override
  Widget build(BuildContext context) {
    final role = context.watch<SessionNotifier>().user?.role;
    final isDiffuseur = role == Role.diffuseur || role == Role.admin;
    final cs = Theme.of(context).colorScheme;
    final bottomPadding = MediaQuery.of(context).padding.bottom;

    final branchMap = isDiffuseur ? [0, 1, 2, 3] : [0, 1, 3];
    final current = navigationShell.currentIndex;
    final tabIndex = branchMap.indexOf(current).clamp(0, branchMap.length - 1);

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const MiniPlayer(),
          Padding(
            padding: EdgeInsets.fromLTRB(16, 0, 16, 12 + bottomPadding),
            child: Container(
              decoration: BoxDecoration(
                color: cs.surfaceContainerHigh,
                borderRadius: BorderRadius.circular(28),
                border: Border.all(color: cs.outlineVariant, width: 0.8),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.3),
                    blurRadius: 24,
                    offset: const Offset(0, 6),
                  ),
                ],
              ),
              child: NavigationBar(
                selectedIndex: tabIndex,
                backgroundColor: Colors.transparent,
                shadowColor: Colors.transparent,
                surfaceTintColor: Colors.transparent,
                elevation: 0,
                onDestinationSelected: (i) => navigationShell.goBranch(
                  branchMap[i],
                  initialLocation: branchMap[i] == current,
                ),
                destinations: isDiffuseur
                    ? const [
                        NavigationDestination(
                          icon: Icon(Icons.home_outlined),
                          selectedIcon: Icon(Icons.home_rounded),
                          label: 'Accueil',
                        ),
                        NavigationDestination(
                          icon: Icon(Icons.radio_outlined),
                          selectedIcon: Icon(Icons.radio_rounded),
                          label: 'Live',
                        ),
                        NavigationDestination(
                          icon: Icon(Icons.mic_none_rounded),
                          selectedIcon: Icon(Icons.mic_rounded),
                          label: 'Studio',
                        ),
                        NavigationDestination(
                          icon: Icon(Icons.library_music_outlined),
                          selectedIcon: Icon(Icons.library_music_rounded),
                          label: 'Bibliothèque',
                        ),
                      ]
                    : const [
                        NavigationDestination(
                          icon: Icon(Icons.home_outlined),
                          selectedIcon: Icon(Icons.home_rounded),
                          label: 'Accueil',
                        ),
                        NavigationDestination(
                          icon: Icon(Icons.radio_outlined),
                          selectedIcon: Icon(Icons.radio_rounded),
                          label: 'Live',
                        ),
                        NavigationDestination(
                          icon: Icon(Icons.library_music_outlined),
                          selectedIcon: Icon(Icons.library_music_rounded),
                          label: 'Bibliothèque',
                        ),
                      ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
