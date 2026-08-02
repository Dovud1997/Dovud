import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final returnsRepositoryProvider = Provider<ReturnsRepository>((ref) {
  return ReturnsRepository(ref.watch(apiClientProvider));
});

class ReturnsRepository {
  ReturnsRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listReturns() async {
    final envelope = await _api.get('/returns', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> createDraft({
    required String customerId,
    required List<Map<String, dynamic>> lines,
    String currency = 'UZS',
    String? reason,
    String? orderId,
  }) async {
    final envelope = await _api.post('/returns', data: {
      'customer_id': customerId,
      'currency': currency,
      'status': 'draft',
      if (reason != null && reason.isNotEmpty) 'reason': reason,
      if (orderId != null && orderId.isNotEmpty) 'order_id': orderId,
      'lines': lines,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }

  Future<Map<String, dynamic>> submit(String returnId) async {
    final envelope = await _api.post('/returns/$returnId/submit');
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }
}
