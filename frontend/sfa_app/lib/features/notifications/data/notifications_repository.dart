import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final notificationsRepositoryProvider = Provider<NotificationsRepository>((ref) {
  return NotificationsRepository(ref.watch(apiClientProvider));
});

final unreadNotificationsCountProvider = FutureProvider.autoDispose<int>((ref) {
  return ref.watch(notificationsRepositoryProvider).unreadCount();
});

class NotificationsRepository {
  NotificationsRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listNotifications({bool unreadOnly = false}) async {
    final envelope = await _api.get('/notifications', query: {
      'page': 1,
      'per_page': 50,
      if (unreadOnly) 'unread': 'true',
    });
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<int> unreadCount() async {
    final envelope = await _api.get('/notifications/unread-count');
    final data = envelope['data'] as Map<String, dynamic>? ?? const {};
    return (data['count'] as num?)?.toInt() ?? 0;
  }

  Future<void> markRead(String notificationId) async {
    await _api.post('/notifications/$notificationId/read');
  }

  Future<int> markAllRead() async {
    final envelope = await _api.post('/notifications/read-all');
    final data = envelope['data'] as Map<String, dynamic>? ?? const {};
    return (data['updated'] as num?)?.toInt() ?? 0;
  }

  Future<List<Map<String, dynamic>>> listDeliveries(String notificationId) async {
    final envelope = await _api.get('/notifications/$notificationId/deliveries');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
