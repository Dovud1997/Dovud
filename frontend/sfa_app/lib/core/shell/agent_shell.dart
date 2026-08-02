import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/notifications/data/notifications_repository.dart';

class AgentShell extends ConsumerWidget {
  const AgentShell({super.key, required this.navigationShell});

  final StatefulNavigationShell navigationShell;

  static const _tabs = <_AgentTab>[
    _AgentTab(label: 'Home', icon: Icons.home_outlined, path: '/home'),
    _AgentTab(label: 'Customers', icon: Icons.people_alt_outlined, path: '/field/customers'),
    _AgentTab(label: 'Orders', icon: Icons.receipt_long_outlined, path: '/field/orders'),
    _AgentTab(label: 'Route', icon: Icons.route_outlined, path: '/field/routes'),
    _AgentTab(label: 'More', icon: Icons.more_horiz, path: '/more'),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final branding = ref.watch(sessionControllerProvider).branding;
    final unread = ref.watch(unreadNotificationsCountProvider).valueOrNull ?? 0;
    return Scaffold(
      appBar: AppBar(
        title: Text(branding?.appName ?? 'SFA'),
        actions: [
          IconButton(
            tooltip: unread > 0 ? 'Notifications ($unread)' : 'Notifications',
            onPressed: () => context.push('/field/notifications'),
            icon: Badge(
              isLabelVisible: unread > 0,
              label: Text(unread > 99 ? '99+' : '$unread'),
              child: const Icon(Icons.notifications_outlined),
            ),
          ),
        ],
      ),
      body: navigationShell,
      bottomNavigationBar: NavigationBar(
        selectedIndex: navigationShell.currentIndex,
        onDestinationSelected: navigationShell.goBranch,
        destinations: [
          for (final t in _tabs)
            NavigationDestination(icon: Icon(t.icon), label: t.label),
        ],
      ),
    );
  }
}

class _AgentTab {
  const _AgentTab({required this.label, required this.icon, required this.path});
  final String label;
  final IconData icon;
  final String path;
}
