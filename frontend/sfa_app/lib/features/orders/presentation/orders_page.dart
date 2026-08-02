import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/orders/data/orders_repository.dart';

final ordersProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(ordersRepositoryProvider).listOrders();
});

class OrdersPage extends ConsumerStatefulWidget {
  const OrdersPage({super.key});

  @override
  ConsumerState<OrdersPage> createState() => _OrdersPageState();
}

class _OrdersPageState extends ConsumerState<OrdersPage> {
  bool _busy = false;

  Future<void> _compose() async {
    final path = GoRouterState.of(context).uri.path;
    final base = path.startsWith('/field') ? '/field/orders' : '/orders';
    final created = await context.push<bool>('$base/new');
    if (created == true) {
      ref.invalidate(ordersProvider);
    }
  }

  Future<void> _submit(String orderId) async {
    setState(() => _busy = true);
    try {
      await ref.read(ordersRepositoryProvider).submit(orderId);
      ref.invalidate(ordersProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(ordersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Orders')),
      floatingActionButton: FloatingActionButton(
        onPressed: _busy ? null : _compose,
        child: const Icon(Icons.add_shopping_cart),
      ),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No orders'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final o = items[i];
                  final status = o['status']?.toString() ?? '';
                  final lines = o['lines'];
                  final lineCount = lines is List ? lines.length : null;
                  return ListTile(
                    title: Text(o['number']?.toString() ?? ''),
                    subtitle: Text(
                      [
                        status,
                        if (lineCount != null) '$lineCount lines',
                        '${o['currency'] ?? ''} ${o['grand_total'] ?? ''}',
                      ].join(' · '),
                    ),
                    trailing: status == 'draft'
                        ? TextButton(
                            onPressed: _busy ? null : () => _submit(o['id']?.toString() ?? ''),
                            child: const Text('Submit'),
                          )
                        : Text(status),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
