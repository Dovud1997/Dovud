import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/auth/presentation/login_page.dart';
import 'package:sfa_app/features/catalog/presentation/products_page.dart';
import 'package:sfa_app/features/crm/presentation/customers_page.dart';
import 'package:sfa_app/features/dashboard/presentation/dashboard_page.dart';
import 'package:sfa_app/features/fieldforce/presentation/routes_page.dart';
import 'package:sfa_app/features/finance/presentation/receivables_page.dart';
import 'package:sfa_app/features/notifications/presentation/notifications_page.dart';
import 'package:sfa_app/features/orders/presentation/orders_page.dart';
import 'package:sfa_app/features/organization/presentation/branches_page.dart';
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
      if (session.isAuthenticated && loggingIn) return '/dashboard';
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/dashboard', builder: (context, state) => const DashboardPage()),
      GoRoute(path: '/products', builder: (context, state) => const ProductsPage()),
      GoRoute(path: '/customers', builder: (context, state) => const CustomersPage()),
      GoRoute(path: '/branches', builder: (context, state) => const BranchesPage()),
      GoRoute(path: '/routes', builder: (context, state) => const RoutesPage()),
      GoRoute(path: '/orders', builder: (context, state) => const OrdersPage()),
      GoRoute(path: '/returns', builder: (context, state) => const ReturnsPage()),
      GoRoute(path: '/receivables', builder: (context, state) => const ReceivablesPage()),
      GoRoute(path: '/notifications', builder: (context, state) => const NotificationsPage()),
      GoRoute(path: '/sync', builder: (context, state) => const SyncPage()),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Text(state.error.toString())),
    ),
  );
});
