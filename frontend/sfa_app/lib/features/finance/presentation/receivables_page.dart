import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/finance/data/finance_repository.dart';

final receivablesProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(financeRepositoryProvider).listReceivables();
});

class ReceivablesPage extends ConsumerWidget {
  const ReceivablesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(receivablesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Receivables')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No receivables'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final r = items[i];
                  return ListTile(
                    title: Text('${r['currency'] ?? ''} ${r['balance'] ?? r['amount'] ?? ''}'),
                    subtitle: Text('${r['status'] ?? ''} · due ${r['due_date'] ?? '—'}'),
                    trailing: Text(r['document_type']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
