import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/fieldforce/data/fieldforce_repository.dart';

final routesProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(fieldForceRepositoryProvider).listRoutes();
});

class RoutesPage extends ConsumerWidget {
  const RoutesPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(routesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Routes')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No routes'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final r = items[i];
                  return ListTile(
                    title: Text(r['name']?.toString() ?? ''),
                    subtitle: Text('${r['date'] ?? ''} · ${r['status'] ?? ''}'),
                    trailing: Text(r['version']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
