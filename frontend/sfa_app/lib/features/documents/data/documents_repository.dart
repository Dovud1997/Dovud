import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final documentsRepositoryProvider = Provider<DocumentsRepository>((ref) {
  return DocumentsRepository(ref.watch(apiClientProvider));
});

class DocumentsRepository {
  DocumentsRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listDocuments() async {
    final envelope = await _api.get('/documents', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<List<Map<String, dynamic>>> listFiles() async {
    final envelope = await _api.get('/files', query: {'page': 1, 'per_page': 50});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
