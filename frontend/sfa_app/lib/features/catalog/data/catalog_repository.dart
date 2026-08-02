import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final catalogRepositoryProvider = Provider<CatalogRepository>((ref) {
  return CatalogRepository(ref.watch(apiClientProvider));
});

class CatalogRepository {
  CatalogRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listProducts({String q = ''}) async {
    final envelope = await _api.get('/products', query: {
      if (q.isNotEmpty) 'q': q,
      'page': 1,
      'per_page': 50,
    });
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listCategories() async {
    final envelope = await _api.get('/categories');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listPriceLists() async {
    final envelope = await _api.get('/price-lists');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listPrices(String priceListId) async {
    if (priceListId.isEmpty) return const [];
    final envelope = await _api.get('/price-lists/$priceListId/prices');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
