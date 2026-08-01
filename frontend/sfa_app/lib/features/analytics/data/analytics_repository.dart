import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final analyticsRepositoryProvider = Provider<AnalyticsRepository>((ref) {
  return AnalyticsRepository(ref.watch(apiClientProvider));
});

class AnalyticsRepository {
  AnalyticsRepository(this._api);
  final ApiClient _api;

  Future<Map<String, dynamic>> dashboardSummary() async {
    final envelope = await _api.get('/dashboard/summary');
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }
}
