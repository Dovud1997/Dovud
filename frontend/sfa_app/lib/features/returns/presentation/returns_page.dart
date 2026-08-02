import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/returns/data/returns_repository.dart';

final returnsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(returnsRepositoryProvider).listReturns();
});

class ReturnsPage extends ConsumerStatefulWidget {
  const ReturnsPage({super.key});

  @override
  ConsumerState<ReturnsPage> createState() => _ReturnsPageState();
}

class _ReturnsPageState extends ConsumerState<ReturnsPage> {
  bool _busy = false;

  Future<void> _compose() async {
    final path = GoRouterState.of(context).uri.path;
    final base = path.startsWith('/field') ? '/field/returns' : '/returns';
    final created = await context.push<bool>('$base/new');
    if (created == true) {
      ref.invalidate(returnsProvider);
    }
  }

  Future<void> _submit(String returnId) async {
    setState(() => _busy = true);
    try {
      await ref.read(returnsRepositoryProvider).submit(returnId);
      ref.invalidate(returnsProvider);
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
    final async = ref.watch(returnsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Returns')),
      floatingActionButton: FloatingActionButton(
        onPressed: _busy ? null : _compose,
        child: const Icon(Icons.assignment_return_outlined),
      ),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No returns'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final r = items[i];
                  final status = r['status']?.toString() ?? '';
                  final lines = r['lines'];
                  final lineCount = lines is List ? lines.length : null;
                  return ListTile(
                    title: Text(r['number']?.toString() ?? ''),
                    subtitle: Text(
                      [
                        status,
                        if (lineCount != null) '$lineCount lines',
                        '${r['currency'] ?? ''} ${r['grand_total'] ?? ''}',
                      ].join(' · '),
                    ),
                    trailing: status == 'draft'
                        ? TextButton(
                            onPressed: _busy ? null : () => _submit(r['id']?.toString() ?? ''),
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
