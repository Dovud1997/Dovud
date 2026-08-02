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

  Future<Map<String, dynamic>> checkIn({
    required String agentId,
    required String customerId,
    double? lat,
    double? lng,
  }) async {
    final envelope = await _api.post('/visits/check-in', data: {
      'agent_id': agentId,
      'customer_id': customerId,
      if (lat != null) 'lat': lat,
      if (lng != null) 'lng': lng,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }

  Future<Map<String, dynamic>> checkOut({
    required String visitId,
    required String result,
  }) async {
    final envelope = await _api.post('/visits/$visitId/check-out', data: {
      'result': result,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map);
  }

  Future<List<Map<String, dynamic>>> uploadGpsPoints(
    List<Map<String, dynamic>> points,
  ) async {
    final envelope = await _api.post('/gps/points', data: points);
    final data = envelope['data'];
    if (data is List) {
      return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
    }
    return const [];
  }
}
