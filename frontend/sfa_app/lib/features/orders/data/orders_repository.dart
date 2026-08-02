import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final ordersRepositoryProvider = Provider<OrdersRepository>((ref) {
  return OrdersRepository(ref.watch(apiClientProvider));
});

class OrdersRepository {
  OrdersRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listOrders() async {
    final envelope = await _api.get('/orders', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> createDraft({
    required String customerId,
    List<Map<String, dynamic>> lines = const [],
    String currency = 'UZS',
    String? clientRequestId,
  }) async {
    final envelope = await _api.post('/orders', data: {
      'customer_id': customerId,
      'currency': currency,
      'status': 'draft',
      if (clientRequestId != null) 'client_request_id': clientRequestId,
      'lines': lines,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }

  Future<Map<String, dynamic>> submit(String orderId) async {
    final envelope = await _api.post('/orders/$orderId/submit');
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }
}
