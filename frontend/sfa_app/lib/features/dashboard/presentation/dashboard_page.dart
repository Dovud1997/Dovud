import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/analytics/data/analytics_repository.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

final dashboardSummaryProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) {
  return ref.watch(analyticsRepositoryProvider).dashboardSummary();
});

class DashboardPage extends ConsumerWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(sessionControllerProvider);
    final user = state.session?.user;
    final theme = Theme.of(context);
    final summary = ref.watch(dashboardSummaryProvider);

    return ListView(
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
        const SizedBox(height: 28),
        Text('Today', style: theme.textTheme.titleLarge),
        const SizedBox(height: 12),
        summary.when(
          data: (s) => Wrap(
            spacing: 12,
            runSpacing: 12,
            children: [
              _KpiChip(label: 'Orders', value: '${s['orders_today'] ?? 0}'),
              _KpiChip(label: 'Visits', value: '${s['visits_today'] ?? 0}'),
              _KpiChip(label: 'Open AR', value: '${s['open_receivables'] ?? 0}'),
              _KpiChip(label: 'Pending', value: '${s['pending_orders'] ?? 0}'),
            ],
          ),
          loading: () => const LinearProgressIndicator(),
          error: (_, __) => Text(
            'KPI unavailable',
            style: theme.textTheme.bodyMedium?.copyWith(
              color: theme.colorScheme.onSurface.withOpacity(0.55),
            ),
          ),
        ),
        const SizedBox(height: 28),
        Text(
          'Use the side navigation to open modules.',
          style: theme.textTheme.bodyMedium?.copyWith(
            color: theme.colorScheme.onSurface.withOpacity(0.65),
          ),
        ),
      ],
    );
  }
}

class _KpiChip extends StatelessWidget {
  const _KpiChip({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: theme.textTheme.labelMedium),
          const SizedBox(height: 4),
          Text(value, style: theme.textTheme.titleMedium),
        ],
      ),
    );
  }
}
