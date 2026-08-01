import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

class DashboardPage extends ConsumerWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(sessionControllerProvider);
    final user = state.session?.user;
    final branding = state.branding;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(branding?.appName ?? 'SFA'),
        actions: [
          IconButton(
            tooltip: 'Log out',
            onPressed: () => ref.read(sessionControllerProvider.notifier).logout(),
            icon: const Icon(Icons.logout_rounded),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          Text(
            'Welcome, ${user?.fullName ?? ''}',
            style: theme.textTheme.headlineMedium,
          ),
          const SizedBox(height: 8),
          Text(
            user?.email ?? '',
            style: theme.textTheme.bodyLarge?.copyWith(
              color: theme.colorScheme.onSurface.withOpacity(0.65),
            ),
          ),
          const SizedBox(height: 24),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final role in user?.roles ?? const <String>[])
                Chip(label: Text(role)),
            ],
          ),
          const SizedBox(height: 24),
          Text('Permissions', style: theme.textTheme.titleMedium),
          const SizedBox(height: 8),
          Text(
            (user?.permissions ?? const []).take(12).join(' · '),
            style: theme.textTheme.bodyMedium,
          ),
          const SizedBox(height: 32),
          Text(
            'P0 shell ready. Next: catalog, CRM, field routes, offline sync.',
            style: theme.textTheme.bodyLarge,
          ),
        ],
      ),
    );
  }
}
