import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/orders/data/orders_repository.dart';

final ordersProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(ordersRepositoryProvider).listOrders();
});

class OrdersPage extends ConsumerWidget {
  const OrdersPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(ordersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Orders')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No orders'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final o = items[i];
                  return ListTile(
                    title: Text(o['number']?.toString() ?? ''),
                    subtitle: Text('${o['status'] ?? ''} · ${o['currency'] ?? ''} ${o['grand_total'] ?? ''}'),
                    trailing: Text(o['status']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
