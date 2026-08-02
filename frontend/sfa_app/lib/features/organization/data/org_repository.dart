import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final orgRepositoryProvider = Provider<OrgRepository>((ref) {
  return OrgRepository(ref.watch(apiClientProvider));
});

class OrgRepository {
  OrgRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listBranches() async {
    final envelope = await _api.get('/branches', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listWarehouses() async {
    final envelope = await _api.get('/warehouses', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
