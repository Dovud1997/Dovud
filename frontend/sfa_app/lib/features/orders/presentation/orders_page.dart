import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/agent_write_coordinator.dart';
import 'package:sfa_app/features/crm/data/crm_repository.dart';
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

  Future<void> _createDraft() async {
    setState(() => _busy = true);
    try {
      final customers = await ref.read(crmRepositoryProvider).listCustomers();
      if (customers.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Create a customer first')),
          );
        }
        return;
      }
      final customerId = customers.first['id']?.toString() ?? '';
      if (customerId.isEmpty) return;
      final reqId = 'flutter-${DateTime.now().microsecondsSinceEpoch}';
      await ref.read(agentWriteCoordinatorProvider).write(
            entityType: 'order',
            op: 'create',
            payload: {
              'customer_id': customerId,
              'currency': 'UZS',
              'status': 'draft',
              'client_request_id': reqId,
              'lines': const <Map<String, dynamic>>[],
            },
            online: () => ref.read(ordersRepositoryProvider).createDraft(
                  customerId: customerId,
                  clientRequestId: reqId,
                ),
          );
      ref.invalidate(ordersProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Draft order created')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
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
        onPressed: _busy ? null : _createDraft,
        child: _busy ? const CircularProgressIndicator() : const Icon(Icons.add_shopping_cart),
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
                  return ListTile(
                    title: Text(o['number']?.toString() ?? ''),
                    subtitle: Text('$status · ${o['currency'] ?? ''} ${o['grand_total'] ?? ''}'),
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
