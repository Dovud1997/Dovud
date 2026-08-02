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

class NotificationsPage extends ConsumerStatefulWidget {
  const NotificationsPage({super.key});

  @override
  ConsumerState<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends ConsumerState<NotificationsPage> {
  bool _busy = false;
  bool _unreadOnly = false;

  Future<void> _markRead(String id) async {
    if (id.isEmpty) return;
    setState(() => _busy = true);
    try {
      await ref.read(notificationsRepositoryProvider).markRead(id);
      ref.invalidate(notificationsProvider);
      ref.invalidate(unreadNotificationsCountProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _markAllRead() async {
    setState(() => _busy = true);
    try {
      final n = await ref.read(notificationsRepositoryProvider).markAllRead();
      ref.invalidate(notificationsProvider);
      ref.invalidate(unreadNotificationsCountProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(n == 0 ? 'Nothing to mark' : 'Marked $n as read')),
        );
      }
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
    final async = ref.watch(notificationsProvider);
    final unreadAsync = ref.watch(unreadNotificationsCountProvider);
    final unread = unreadAsync.valueOrNull ?? 0;
    return Scaffold(
      appBar: AppBar(
        title: Text(unread > 0 ? 'Notifications ($unread)' : 'Notifications'),
        actions: [
          IconButton(
            tooltip: _unreadOnly ? 'Show all' : 'Unread only',
            onPressed: _busy
                ? null
                : () {
                    setState(() => _unreadOnly = !_unreadOnly);
                    // filter client-side from cached list
                  },
            icon: Icon(_unreadOnly ? Icons.filter_alt : Icons.filter_alt_outlined),
          ),
          TextButton(
            onPressed: _busy || unread == 0 ? null : _markAllRead,
            child: const Text('Mark all'),
          ),
        ],
      ),
      body: async.when(
        data: (items) {
          final filtered = _unreadOnly
              ? items.where((n) => n['read_at'] == null).toList()
              : items;
          if (filtered.isEmpty) {
            return Center(child: Text(_unreadOnly ? 'No unread' : 'No notifications'));
          }
          return ListView.separated(
            itemCount: filtered.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, i) {
              final n = filtered[i];
              final unreadItem = n['read_at'] == null;
              final id = n['id']?.toString() ?? '';
              return ExpansionTile(
                leading: Icon(
                  unreadItem ? Icons.mark_email_unread_outlined : Icons.mark_email_read_outlined,
                  color: unreadItem ? Theme.of(context).colorScheme.primary : null,
                ),
                title: Text(
                  n['title']?.toString() ?? '',
                  style: TextStyle(fontWeight: unreadItem ? FontWeight.w600 : FontWeight.w400),
                ),
                subtitle: Text('${n['body'] ?? ''} · ${n['channel'] ?? ''}'),
                trailing: unreadItem
                    ? TextButton(
                        onPressed: _busy ? null : () => _markRead(id),
                        child: const Text('Read'),
                      )
                    : null,
                onExpansionChanged: (open) {
                  if (open && unreadItem) {
                    _markRead(id);
                  }
                },
                children: [
                  if (id.isEmpty)
                    const ListTile(title: Text('Missing id'))
                  else
                    _DeliveriesList(notificationId: id),
                ],
              );
            },
          );
        },
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
