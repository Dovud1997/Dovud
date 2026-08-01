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
}
