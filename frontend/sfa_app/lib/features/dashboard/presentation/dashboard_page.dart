import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
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
    final branding = state.branding;
    final theme = Theme.of(context);
    final summary = ref.watch(dashboardSummaryProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(branding?.appName ?? 'SFA'),
        actions: [
          IconButton(
            tooltip: 'Notifications',
            onPressed: () => context.push('/notifications'),
            icon: const Icon(Icons.notifications_outlined),
          ),
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
          Text('Modules', style: theme.textTheme.titleLarge),
          const SizedBox(height: 12),
          _ModuleTile(
            icon: Icons.store_mall_directory_outlined,
            title: 'Branches',
            subtitle: 'Organization structure',
            onTap: () => context.push('/branches'),
          ),
          _ModuleTile(
            icon: Icons.inventory_2_outlined,
            title: 'Products',
            subtitle: 'Catalog & SKUs',
            onTap: () => context.push('/products'),
          ),
          _ModuleTile(
            icon: Icons.people_alt_outlined,
            title: 'Customers',
            subtitle: 'CRM outlets & contacts',
            onTap: () => context.push('/customers'),
          ),
          _ModuleTile(
            icon: Icons.route_outlined,
            title: 'Routes',
            subtitle: 'Field visits & stops',
            onTap: () => context.push('/routes'),
          ),
          _ModuleTile(
            icon: Icons.receipt_long_outlined,
            title: 'Orders',
            subtitle: 'Sales orders pipeline',
            onTap: () => context.push('/orders'),
          ),
          _ModuleTile(
            icon: Icons.assignment_return_outlined,
            title: 'Returns',
            subtitle: 'Return approvals',
            onTap: () => context.push('/returns'),
          ),
          _ModuleTile(
            icon: Icons.account_balance_wallet_outlined,
            title: 'Receivables',
            subtitle: 'Finance / AR',
            onTap: () => context.push('/receivables'),
          ),
          _ModuleTile(
            icon: Icons.folder_outlined,
            title: 'Documents',
            subtitle: 'Files & document meta',
            onTap: () => context.push('/documents'),
          ),
          _ModuleTile(
            icon: Icons.policy_outlined,
            title: 'Audit logs',
            subtitle: 'Security & change trail',
            onTap: () => context.push('/audit'),
          ),
          _ModuleTile(
            icon: Icons.storefront_outlined,
            title: 'Customer portal',
            subtitle: 'Read-only B2B customer view',
            onTap: () => context.push('/portal'),
          ),
          _ModuleTile(
            icon: Icons.sync_outlined,
            title: 'Sync center',
            subtitle: 'Offline outbox · pull / push',
            onTap: () => context.push('/sync'),
          ),
        ],
      ),
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

class _ModuleTile extends StatelessWidget {
  const _ModuleTile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: ListTile(
        onTap: onTap,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        tileColor: theme.colorScheme.surface,
        leading: Icon(icon, color: theme.colorScheme.primary),
        title: Text(title),
        subtitle: Text(subtitle),
        trailing: const Icon(Icons.chevron_right),
      ),
    );
  }
}
