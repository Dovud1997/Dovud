import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/organization/data/org_repository.dart';

final branchesProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(orgRepositoryProvider).listBranches();
});

class BranchesPage extends ConsumerWidget {
  const BranchesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(branchesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Branches')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No branches'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final b = items[i];
                  return ListTile(
                    title: Text(b['name']?.toString() ?? ''),
                    subtitle: Text('${b['code'] ?? ''} · ${b['address'] ?? ''}'),
                    trailing: Text(b['status']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
