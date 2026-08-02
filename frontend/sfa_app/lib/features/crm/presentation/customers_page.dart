import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/agent_write_coordinator.dart';
import 'package:sfa_app/features/crm/data/crm_repository.dart';

final customersProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(crmRepositoryProvider).listCustomers();
});

class CustomersPage extends ConsumerStatefulWidget {
  const CustomersPage({super.key});

  @override
  ConsumerState<CustomersPage> createState() => _CustomersPageState();
}

class _CustomersPageState extends ConsumerState<CustomersPage> {
  bool _busy = false;

  Future<void> _createCustomer() async {
    final codeCtrl = TextEditingController();
    final nameCtrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New customer'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: codeCtrl, decoration: const InputDecoration(labelText: 'Code')),
            TextField(controller: nameCtrl, decoration: const InputDecoration(labelText: 'Name')),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Save')),
        ],
      ),
    );
    if (ok != true || !mounted) return;
    final code = codeCtrl.text.trim();
    final name = nameCtrl.text.trim();
    if (code.isEmpty || name.isEmpty) return;

    setState(() => _busy = true);
    try {
      await ref.read(agentWriteCoordinatorProvider).write(
            entityType: 'customer',
            op: 'create',
            payload: {'code': code, 'name': name, 'type': 'outlet', 'status': 'active'},
            online: () => ref.read(crmRepositoryProvider).createCustomer(code: code, name: name),
          );
      ref.invalidate(customersProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Customer saved')));
      }
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
    final async = ref.watch(customersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Customers')),
      floatingActionButton: FloatingActionButton(
        onPressed: _busy ? null : _createCustomer,
        child: _busy ? const CircularProgressIndicator() : const Icon(Icons.person_add_alt_1),
      ),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No customers'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final c = items[i];
                  return ListTile(
                    title: Text(c['name']?.toString() ?? ''),
                    subtitle: Text('${c['code'] ?? ''} · ${c['type'] ?? ''}'),
                    trailing: Text(c['status']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
