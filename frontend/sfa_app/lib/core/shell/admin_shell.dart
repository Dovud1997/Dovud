import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/core/shell/nav_destinations.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/notifications/data/notifications_repository.dart';

class AdminShell extends ConsumerWidget {
  const AdminShell({super.key, required this.navigationShell});

  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(sessionControllerProvider);
    final user = session.session?.user;
    final branding = session.branding;
    final destinations = adminDestinationsFor(user);
    final wide = MediaQuery.sizeOf(context).width >= 900;
    final idx = _selectedIndex(destinations, GoRouterState.of(context).uri.path);
    final unread = ref.watch(unreadNotificationsCountProvider).valueOrNull ?? 0;

    final rail = NavigationRail(
      selectedIndex: idx < 0 ? 0 : idx,
      labelType: NavigationRailLabelType.all,
      onDestinationSelected: (i) {
        final dest = destinations[i];
        if (dest.branchIndex >= 0) {
          navigationShell.goBranch(dest.branchIndex);
        } else {
          context.go(dest.path);
        }
      },
      destinations: [
        for (final d in destinations)
          NavigationRailDestination(icon: Icon(d.icon), label: Text(d.label)),
      ],
    );

    final body = Row(
      children: [
        if (wide) ...[
          rail,
          const VerticalDivider(width: 1),
        ],
        Expanded(child: navigationShell),
      ],
    );

    return Scaffold(
      appBar: AppBar(
        title: Text(branding?.appName ?? 'SFA'),
        leading: wide
            ? null
            : Builder(
                builder: (ctx) => IconButton(
                  icon: const Icon(Icons.menu),
                  onPressed: () => Scaffold.of(ctx).openDrawer(),
                ),
              ),
        actions: [
          IconButton(
            tooltip: unread > 0 ? 'Notifications ($unread)' : 'Notifications',
            onPressed: () => context.go('/notifications'),
            icon: Badge(
              isLabelVisible: unread > 0,
              label: Text(unread > 99 ? '99+' : '$unread'),
              child: const Icon(Icons.notifications_outlined),
            ),
          ),
          IconButton(
            tooltip: 'Log out',
            onPressed: () => ref.read(sessionControllerProvider.notifier).logout(),
            icon: const Icon(Icons.logout_rounded),
          ),
        ],
      ),
      drawer: wide
          ? null
          : Drawer(
              child: ListView(
                children: [
                  DrawerHeader(
                    child: Text(
                      branding?.appName ?? 'SFA',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                  ),
                  for (var i = 0; i < destinations.length; i++)
                    ListTile(
                      leading: Icon(destinations[i].icon),
                      title: Text(destinations[i].label),
                      selected: i == idx,
                      onTap: () {
                        Navigator.pop(context);
                        final dest = destinations[i];
                        if (dest.branchIndex >= 0) {
                          navigationShell.goBranch(dest.branchIndex);
                        } else {
                          context.go(dest.path);
                        }
                      },
                    ),
                ],
              ),
            ),
      body: body,
    );
  }

  int _selectedIndex(List<AdminDestination> destinations, String path) {
    for (var i = 0; i < destinations.length; i++) {
      if (path == destinations[i].path || path.startsWith('${destinations[i].path}/')) {
        return i;
      }
    }
    return 0;
  }
}
