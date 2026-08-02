import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final crmRepositoryProvider = Provider<CrmRepository>((ref) {
  return CrmRepository(ref.watch(apiClientProvider));
});

class CrmRepository {
  CrmRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listCustomers({String q = ''}) async {
    final envelope = await _api.get('/customers', query: {
      if (q.isNotEmpty) 'q': q,
      'page': 1,
      'per_page': 50,
    });
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> createCustomer({
    required String code,
    required String name,
    String type = 'outlet',
  }) async {
    final envelope = await _api.post('/customers', data: {
      'code': code,
      'name': name,
      'type': type,
      'status': 'active',
    });
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }
}
