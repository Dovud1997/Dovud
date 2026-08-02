import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/notifications/data/notifications_repository.dart';

final notificationsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(notificationsRepositoryProvider).listNotifications();
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
                  return ListTile(
                    leading: Icon(
                      unread ? Icons.mark_email_unread_outlined : Icons.mark_email_read_outlined,
                    ),
                    title: Text(n['title']?.toString() ?? ''),
                    subtitle: Text(n['body']?.toString() ?? ''),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('$e')),
      ),
    );
  }
}
