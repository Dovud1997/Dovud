import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final syncStatusProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) {
  return ref.watch(syncRepositoryProvider).status();
});

class SyncPage extends ConsumerWidget {
  const SyncPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(syncStatusProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sync center'),
        actions: [
          IconButton(
            tooltip: 'Bootstrap device',
            onPressed: () async {
              await ref.read(syncRepositoryProvider).bootstrap();
              ref.invalidate(syncStatusProvider);
            },
            icon: const Icon(Icons.cloud_sync_outlined),
          ),
        ],
      ),
      body: async.when(
        data: (s) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            ListTile(
              title: const Text('Device'),
              subtitle: Text(s['device_id']?.toString() ?? '—'),
            ),
            ListTile(
              title: const Text('Protocol'),
              subtitle: Text(s['sync_protocol']?.toString() ?? '—'),
            ),
            ListTile(
              title: const Text('Last pull cursor'),
              subtitle: Text(s['last_pull_cursor']?.toString().isEmpty == true
                  ? '—'
                  : s['last_pull_cursor']?.toString() ?? '—'),
            ),
            ListTile(
              title: const Text('Open conflicts'),
              subtitle: Text(s['open_conflicts']?.toString() ?? '0'),
            ),
          ],
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Text('$e\n\nTap cloud icon to bootstrap this device.'),
          ),
        ),
      ),
    );
  }
}
