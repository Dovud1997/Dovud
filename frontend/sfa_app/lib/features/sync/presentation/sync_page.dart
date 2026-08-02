import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final syncStatusProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) {
  return ref.watch(syncRepositoryProvider).status();
});

final outboxCountProvider = FutureProvider.autoDispose<int>((ref) async {
  return (await ref.watch(localOutboxProvider).list()).length;
});

class SyncPage extends ConsumerStatefulWidget {
  const SyncPage({super.key});

  @override
  ConsumerState<SyncPage> createState() => _SyncPageState();
}

class _SyncPageState extends ConsumerState<SyncPage> {
  String _lastPullSummary = '';
  String _lastFlushSummary = '';
  bool _busy = false;

  Future<void> _run(Future<void> Function() action) async {
    setState(() => _busy = true);
    try {
      await action();
      ref.invalidate(syncStatusProvider);
      ref.invalidate(outboxCountProvider);
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
    final async = ref.watch(syncStatusProvider);
    final outboxCount = ref.watch(outboxCountProvider).valueOrNull ?? 0;
    return Scaffold(
      appBar: AppBar(title: const Text('Sync center')),
      body: Column(
        children: [
          if (_busy) const LinearProgressIndicator(),
          Expanded(
            child: async.when(
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
                    subtitle: Text(
                      (s['last_pull_cursor']?.toString().isEmpty ?? true)
                          ? '—'
                          : s['last_pull_cursor'].toString(),
                    ),
                  ),
                  ListTile(
                    title: const Text('Open conflicts'),
                    subtitle: Text(s['open_conflicts']?.toString() ?? '0'),
                  ),
                  ListTile(
                    title: const Text('Local outbox'),
                    subtitle: Text('$outboxCount pending ops'),
                  ),
                  if (_lastPullSummary.isNotEmpty)
                    ListTile(
                      title: const Text('Last pull'),
                      subtitle: Text(_lastPullSummary),
                    ),
                  if (_lastFlushSummary.isNotEmpty)
                    ListTile(
                      title: const Text('Last flush'),
                      subtitle: Text(_lastFlushSummary),
                    ),
                  const SizedBox(height: 16),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      FilledButton.icon(
                        onPressed: _busy
                            ? null
                            : () => _run(() async {
                                  await ref.read(syncRepositoryProvider).bootstrap();
                                }),
                        icon: const Icon(Icons.cloud_sync_outlined),
                        label: const Text('Bootstrap'),
                      ),
                      FilledButton.tonalIcon(
                        onPressed: _busy
                            ? null
                            : () => _run(() async {
                                  final cursor = s['last_pull_cursor']?.toString() ?? '';
                                  final res = await ref.read(syncRepositoryProvider).pull(cursor: cursor);
                                  final changes = (res['changes'] as List?)?.length ?? 0;
                                  setState(() => _lastPullSummary = '$changes changes');
                                }),
                        icon: const Icon(Icons.download_outlined),
                        label: const Text('Pull'),
                      ),
                      OutlinedButton.icon(
                        onPressed: _busy
                            ? null
                            : () => _run(() async {
                                  final id = DateTime.now().microsecondsSinceEpoch.toString();
                                  await ref.read(localOutboxProvider).enqueue(
                                        OutboxOp(
                                          opId: 'local-$id',
                                          entityType: 'note',
                                          entityId: 'note-$id',
                                          op: 'create',
                                          payload: {'text': 'Offline note $id'},
                                        ),
                                      );
                                }),
                        icon: const Icon(Icons.add_box_outlined),
                        label: const Text('Queue offline op'),
                      ),
                      FilledButton.icon(
                        onPressed: _busy
                            ? null
                            : () => _run(() async {
                                  final res = await ref.read(localOutboxProvider).flush();
                                  setState(() {
                                    _lastFlushSummary =
                                        'acked ${res['acked']}, conflicts ${res['conflicts']}, remaining ${res['remaining']}';
                                  });
                                }),
                        icon: const Icon(Icons.upload_outlined),
                        label: const Text('Flush outbox'),
                      ),
                    ],
                  ),
                ],
              ),
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text('$e'),
                      const SizedBox(height: 12),
                      FilledButton(
                        onPressed: () => _run(() async {
                          await ref.read(syncRepositoryProvider).bootstrap();
                        }),
                        child: const Text('Bootstrap device'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
