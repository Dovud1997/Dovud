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
}
