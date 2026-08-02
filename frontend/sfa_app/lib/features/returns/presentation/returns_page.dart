import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/returns/data/returns_repository.dart';

final returnsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(returnsRepositoryProvider).listReturns();
});

class ReturnsPage extends ConsumerWidget {
  const ReturnsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(returnsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Returns')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No returns'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final r = items[i];
                  return ListTile(
                    title: Text(r['number']?.toString() ?? ''),
                    subtitle: Text('${r['status'] ?? ''} · ${r['currency'] ?? ''} ${r['grand_total'] ?? ''}'),
                    trailing: Text(r['status']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
