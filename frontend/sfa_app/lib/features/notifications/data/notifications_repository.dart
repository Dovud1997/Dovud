import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final notificationsRepositoryProvider = Provider<NotificationsRepository>((ref) {
  return NotificationsRepository(ref.watch(apiClientProvider));
});

class NotificationsRepository {
  NotificationsRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listNotifications() async {
    final envelope = await _api.get('/notifications', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<int> unreadCount() async {
    final envelope = await _api.get('/notifications/unread-count');
    final data = envelope['data'] as Map<String, dynamic>? ?? const {};
    return (data['count'] as num?)?.toInt() ?? 0;
  }
}
