import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final domainsRepositoryProvider = Provider<DomainsRepository>((ref) {
  return DomainsRepository(ref.watch(apiClientProvider));
});

class DomainsRepository {
  DomainsRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> list() async {
    final envelope = await _api.get('/tenant/domains');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> add({required String host, bool isPrimary = false}) async {
    final envelope = await _api.post('/tenant/domains', data: {
      'host': host,
      'is_primary': isPrimary,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<void> delete(String id) async {
    await _api.delete('/tenant/domains/$id');
  }
}
