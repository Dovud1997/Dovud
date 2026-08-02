import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/documents/data/documents_repository.dart';

final documentsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(documentsRepositoryProvider).listDocuments();
});

class DocumentsPage extends ConsumerWidget {
  const DocumentsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(documentsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Documents')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No documents'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final d = items[i];
                  return ListTile(
                    title: Text(d['title']?.toString() ?? ''),
                    subtitle: Text('${d['doc_type'] ?? ''} · ${d['status'] ?? ''}'),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
