import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/catalog/data/catalog_repository.dart';

final productsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(catalogRepositoryProvider).listProducts();
});

class ProductsPage extends ConsumerWidget {
  const ProductsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(productsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Products')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No products'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final p = items[i];
                  return ListTile(
                    title: Text(p['name']?.toString() ?? ''),
                    subtitle: Text('${p['sku'] ?? ''} · ${p['unit'] ?? ''}'),
                    trailing: Text(p['is_active'] == true ? 'Active' : 'Off'),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
