import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final fieldForceRepositoryProvider = Provider<FieldForceRepository>((ref) {
  return FieldForceRepository(ref.watch(apiClientProvider));
});

class FieldForceRepository {
  FieldForceRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listRoutes() async {
    final envelope = await _api.get('/routes', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listVisits() async {
    final envelope = await _api.get('/visits', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listAgents() async {
    final envelope = await _api.get('/agents', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
