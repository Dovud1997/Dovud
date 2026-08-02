import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/notifications/data/notifications_repository.dart';

final notificationsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(notificationsRepositoryProvider).listNotifications();
});

final deliveriesProvider =
    FutureProvider.autoDispose.family<List<Map<String, dynamic>>, String>((ref, id) {
  return ref.watch(notificationsRepositoryProvider).listDeliveries(id);
});

class NotificationsPage extends ConsumerWidget {
  const NotificationsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(notificationsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Notifications')),
      body: async.when(
        data: (items) => items.isEmpty
            ? const Center(child: Text('No notifications'))
            : ListView.separated(
                itemCount: items.length,
                separatorBuilder: (_, __) => const Divider(height: 1),
                itemBuilder: (context, i) {
                  final n = items[i];
                  final unread = n['read_at'] == null;
                  final id = n['id']?.toString() ?? '';
                  return ExpansionTile(
                    leading: Icon(
                      unread ? Icons.mark_email_unread_outlined : Icons.mark_email_read_outlined,
                    ),
                    title: Text(n['title']?.toString() ?? ''),
                    subtitle: Text('${n['body'] ?? ''} · ${n['channel'] ?? ''}'),
                    children: [
                      if (id.isEmpty)
                        const ListTile(title: Text('Missing id'))
                      else
                        _DeliveriesList(notificationId: id),
                    ],
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}

class _DeliveriesList extends ConsumerWidget {
  const _DeliveriesList({required this.notificationId});

  final String notificationId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(deliveriesProvider(notificationId));
    return async.when(
      data: (rows) {
        if (rows.isEmpty) {
          return const ListTile(dense: true, title: Text('No delivery attempts yet'));
        }
        return Column(
          children: [
            for (final d in rows)
              ListTile(
                dense: true,
                title: Text(
                  '${d['channel'] ?? ''} · ${d['status'] ?? ''}'
                  '${d['device_id'] != null ? ' · ${d['device_id']}' : ''}',
                ),
                subtitle: Text(
                  [
                    if (d['platform'] != null) 'platform=${d['platform']}',
                    if (d['token_suffix'] != null) '…${d['token_suffix']}',
                    if (d['error'] != null) d['error'],
                  ].where((e) => e.toString().isNotEmpty).join(' · '),
                ),
              ),
          ],
        );
      },
      loading: () => const Padding(
        padding: EdgeInsets.all(12),
        child: LinearProgressIndicator(),
      ),
      error: (e, _) => ListTile(dense: true, title: Text('$e')),
    );
  }
}
