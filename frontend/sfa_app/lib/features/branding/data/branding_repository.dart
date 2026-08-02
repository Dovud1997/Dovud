import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';
import 'package:sfa_app/core/theme/brand_theme.dart';

final brandingRepositoryProvider = Provider<BrandingRepository>((ref) {
  return BrandingRepository(ref.watch(apiClientProvider));
});

class BrandingRepository {
  BrandingRepository(this._api);

  final ApiClient _api;

  Future<Branding> fetchPublic({required String tenantCode}) async {
    final envelope = await _api.get('/public/branding', query: {'tenant': tenantCode});
    final data = Map<String, dynamic>.from(envelope['data'] as Map);
    return Branding.fromJson(data);
  }

  Future<Map<String, dynamic>> fetchTenant() async {
    final envelope = await _api.get('/tenant/branding');
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> updateTenant(Map<String, dynamic> body) async {
    final envelope = await _api.put('/tenant/branding', data: body);
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> attachAsset({
    required String fileId,
    required String kind,
  }) async {
    final envelope = await _api.post('/tenant/branding/assets', data: {
      'file_id': fileId,
      'kind': kind,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }
}
