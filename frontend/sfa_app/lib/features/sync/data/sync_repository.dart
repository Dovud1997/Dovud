import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final syncRepositoryProvider = Provider<SyncRepository>((ref) {
  return SyncRepository(ref.watch(apiClientProvider));
});

class SyncRepository {
  SyncRepository(this._api);
  final ApiClient _api;

  Future<Map<String, dynamic>> status({String deviceId = 'flutter-web'}) async {
    final envelope = await _api.get('/sync/status', query: {'device_id': deviceId});
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> bootstrap({String deviceId = 'flutter-web'}) async {
    final envelope = await _api.post('/sync/bootstrap', data: {
      'device_id': deviceId,
      'platform': 'web',
      'app_version': '0.1.0',
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }
}
