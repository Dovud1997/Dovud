import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final providersRepositoryProvider = Provider<ProvidersRepository>((ref) {
  return ProvidersRepository(ref.watch(apiClientProvider));
});

class ProvidersRepository {
  ProvidersRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> list() async {
    final envelope = await _api.get('/tenant/providers');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> upsert(String type, Map<String, dynamic> body) async {
    final envelope = await _api.put('/tenant/providers/$type', data: body);
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<void> test(String type, {String? to}) async {
    await _api.post('/tenant/providers/$type/test', data: {
      if (to != null && to.isNotEmpty) 'to': to,
    });
  }
}
