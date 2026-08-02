import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final portalRepositoryProvider = Provider<PortalRepository>((ref) {
  return PortalRepository(ref.watch(apiClientProvider));
});

class PortalRepository {
  PortalRepository(this._api);
  final ApiClient _api;

  Future<Map<String, dynamic>> summary() async {
    final envelope = await _api.get('/portal/summary');
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<List<Map<String, dynamic>>> orders() async {
    final envelope = await _api.get('/portal/orders', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> receivables() async {
    final envelope = await _api.get('/portal/receivables', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
