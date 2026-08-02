import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/audit/data/audit_repository.dart';

final auditLogsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(auditRepositoryProvider).listLogs();
});

class AuditPage extends ConsumerWidget {
  const AuditPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(auditLogsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Audit logs')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No audit events yet'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final a = items[i];
                  return ListTile(
                    title: Text(a['action']?.toString() ?? ''),
                    subtitle: Text(
                      '${a['entity_type'] ?? '—'} · ${a['created_at'] ?? ''}',
                    ),
                    trailing: Text(a['request_id']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
