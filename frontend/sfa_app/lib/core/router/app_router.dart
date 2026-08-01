import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/auth/presentation/login_page.dart';
import 'package:sfa_app/features/dashboard/presentation/dashboard_page.dart';

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
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Text(state.error.toString())),
    ),
  );
});
