import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/core/shell/admin_shell.dart';
import 'package:sfa_app/core/shell/agent_shell.dart';
import 'package:sfa_app/features/agent/presentation/agent_more_page.dart';
import 'package:sfa_app/features/audit/presentation/audit_page.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/auth/presentation/login_page.dart';
import 'package:sfa_app/features/branding/presentation/branding_page.dart';
import 'package:sfa_app/features/catalog/presentation/products_page.dart';
import 'package:sfa_app/features/crm/presentation/customers_page.dart';
import 'package:sfa_app/features/dashboard/presentation/dashboard_page.dart';
import 'package:sfa_app/features/documents/presentation/documents_page.dart';
import 'package:sfa_app/features/fieldforce/presentation/routes_page.dart';
import 'package:sfa_app/features/finance/presentation/receivables_page.dart';
import 'package:sfa_app/features/identity/presentation/roles_page.dart';
import 'package:sfa_app/features/identity/presentation/users_page.dart';
import 'package:sfa_app/features/notifications/presentation/notifications_page.dart';
import 'package:sfa_app/features/orders/presentation/order_compose_page.dart';
import 'package:sfa_app/features/orders/presentation/orders_page.dart';
import 'package:sfa_app/features/organization/presentation/branches_page.dart';
import 'package:sfa_app/features/portal/presentation/portal_links_page.dart';
import 'package:sfa_app/features/portal/presentation/portal_page.dart';
import 'package:sfa_app/features/returns/presentation/returns_page.dart';
import 'package:sfa_app/features/sync/presentation/conflict_resolve_page.dart';
import 'package:sfa_app/features/sync/presentation/sync_page.dart';
import 'package:sfa_app/features/tenant/presentation/domains_page.dart';
import 'package:sfa_app/features/tenant/presentation/providers_page.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final session = ref.watch(sessionControllerProvider);

  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      if (session.loading) return null;
      final loggingIn = state.matchedLocation == '/login';
      if (!session.isAuthenticated && !loggingIn) return '/login';
      final user = session.session?.user;
      final isPortal = user?.isPortal ?? false;
      final isAgent = user?.isAgent ?? false;
      final loc = state.matchedLocation;

      if (session.isAuthenticated && loggingIn) {
        if (isPortal) return '/portal';
        if (isAgent) return '/home';
        return '/dashboard';
      }
      if (session.isAuthenticated && isPortal) {
        if (loc != '/portal' && loc != '/login') return '/portal';
      }
      if (session.isAuthenticated && isAgent) {
        final ok = loc == '/home' ||
            loc.startsWith('/field/') ||
            loc == '/more' ||
            loc.startsWith('/more/') ||
            loc == '/login';
        if (!ok) return '/home';
      }
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/portal', builder: (context, state) => const PortalPage()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) => AgentShell(navigationShell: navigationShell),
        branches: [
          StatefulShellBranch(routes: [
            GoRoute(path: '/home', builder: (context, state) => const DashboardPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/field/customers', builder: (context, state) => const CustomersPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/field/orders',
              builder: (context, state) => const OrdersPage(),
              routes: [
                GoRoute(path: 'new', builder: (context, state) => const OrderComposePage()),
              ],
            ),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/field/routes', builder: (context, state) => const RoutesPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/more',
              builder: (context, state) => const AgentMorePage(),
              routes: [
                GoRoute(path: 'sync', builder: (context, state) => const SyncPage()),
                GoRoute(path: 'notifications', builder: (context, state) => const NotificationsPage()),
              ],
            ),
          ]),
        ],
      ),
      // nested under /more as /more/sync — agent more page uses /field/sync; fix paths
      GoRoute(
        path: '/field/sync',
        builder: (context, state) => const SyncPage(),
        routes: [
          GoRoute(
            path: 'conflicts/:id',
            builder: (context, state) => ConflictResolvePage(
              conflictId: state.pathParameters['id'] ?? '',
            ),
          ),
        ],
      ),
      GoRoute(path: '/field/notifications', builder: (context, state) => const NotificationsPage()),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) => AdminShell(navigationShell: navigationShell),
        branches: [
          StatefulShellBranch(routes: [
            GoRoute(path: '/dashboard', builder: (context, state) => const DashboardPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/branches', builder: (context, state) => const BranchesPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/products', builder: (context, state) => const ProductsPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/customers', builder: (context, state) => const CustomersPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/routes', builder: (context, state) => const RoutesPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/orders',
              builder: (context, state) => const OrdersPage(),
              routes: [
                GoRoute(path: 'new', builder: (context, state) => const OrderComposePage()),
              ],
            ),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/returns', builder: (context, state) => const ReturnsPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/receivables', builder: (context, state) => const ReceivablesPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/documents', builder: (context, state) => const DocumentsPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/audit', builder: (context, state) => const AuditPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/providers', builder: (context, state) => const ProvidersPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/portal-links', builder: (context, state) => const PortalLinksPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/users', builder: (context, state) => const UsersPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/roles', builder: (context, state) => const RolesPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/branding', builder: (context, state) => const BrandingPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/domains', builder: (context, state) => const DomainsPage()),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(
              path: '/sync',
              builder: (context, state) => const SyncPage(),
              routes: [
                GoRoute(
                  path: 'conflicts/:id',
                  builder: (context, state) => ConflictResolvePage(
                    conflictId: state.pathParameters['id'] ?? '',
                  ),
                ),
              ],
            ),
          ]),
          StatefulShellBranch(routes: [
            GoRoute(path: '/notifications', builder: (context, state) => const NotificationsPage()),
          ]),
        ],
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Text(state.error.toString())),
    ),
  );
});
