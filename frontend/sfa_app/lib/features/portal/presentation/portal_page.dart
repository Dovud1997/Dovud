import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/portal/data/portal_repository.dart';

final portalSummaryProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) {
  return ref.watch(portalRepositoryProvider).summary();
});

final portalOrdersProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(portalRepositoryProvider).orders();
});

class PortalPage extends ConsumerWidget {
  const PortalPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final summary = ref.watch(portalSummaryProvider);
    final orders = ref.watch(portalOrdersProvider);
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Customer portal')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          summary.when(
            data: (s) => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(s['customer_name']?.toString() ?? 'Customer', style: theme.textTheme.headlineSmall),
                const SizedBox(height: 4),
                Text('${s['customer_code'] ?? ''}'),
                const SizedBox(height: 16),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: [
                    _Stat(label: 'Open orders', value: '${s['open_orders'] ?? 0}'),
                    _Stat(label: 'Balance', value: '${s['open_balance'] ?? 0}'),
                    _Stat(label: 'Credit limit', value: '${s['credit_limit'] ?? 0}'),
                  ],
                ),
              ],
            ),
            loading: () => const LinearProgressIndicator(),
            error: (e, _) => Text('$e'),
          ),
          const SizedBox(height: 28),
          Text('Orders', style: theme.textTheme.titleLarge),
          const SizedBox(height: 8),
          orders.when(
            data: (items) => items.isEmpty
                ? const Text('No orders')
                : Column(
                    children: [
                      for (final o in items)
                        ListTile(
                          contentPadding: EdgeInsets.zero,
                          title: Text(o['number']?.toString() ?? ''),
                          subtitle: Text('${o['status'] ?? ''} · ${o['currency'] ?? ''} ${o['grand_total'] ?? ''}'),
                        ),
                    ],
                  ),
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Text('$e'),
          ),
        ],
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.label, required this.value});
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
