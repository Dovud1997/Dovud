import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/core/shell/admin_shell.dart';
import 'package:sfa_app/features/audit/presentation/audit_page.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/auth/presentation/login_page.dart';
import 'package:sfa_app/features/catalog/presentation/products_page.dart';
import 'package:sfa_app/features/crm/presentation/customers_page.dart';
import 'package:sfa_app/features/dashboard/presentation/dashboard_page.dart';
import 'package:sfa_app/features/documents/presentation/documents_page.dart';
import 'package:sfa_app/features/fieldforce/presentation/routes_page.dart';
import 'package:sfa_app/features/finance/presentation/receivables_page.dart';
import 'package:sfa_app/features/notifications/presentation/notifications_page.dart';
import 'package:sfa_app/features/orders/presentation/orders_page.dart';
import 'package:sfa_app/features/organization/presentation/branches_page.dart';
import 'package:sfa_app/features/portal/presentation/portal_page.dart';
import 'package:sfa_app/features/returns/presentation/returns_page.dart';
import 'package:sfa_app/features/sync/presentation/sync_page.dart';

final appRouterProvider = Provider<GoRouter>((ref) {
  final session = ref.watch(sessionControllerProvider);

  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      if (session.loading) return null;
      final loggingIn = state.matchedLocation == '/login';
      if (!session.isAuthenticated && !loggingIn) return '/login';
      final isPortal = session.session?.user.isPortal ?? false;
      if (session.isAuthenticated && loggingIn) {
        return isPortal ? '/portal' : '/dashboard';
      }
      if (session.isAuthenticated && isPortal) {
        final loc = state.matchedLocation;
        if (loc != '/portal' && loc != '/login') return '/portal';
      }
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/portal', builder: (context, state) => const PortalPage()),
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
            GoRoute(path: '/orders', builder: (context, state) => const OrdersPage()),
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
            GoRoute(path: '/sync', builder: (context, state) => const SyncPage()),
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
