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

  Future<Map<String, dynamic>> presign({
    required String fileName,
    required String mime,
    required int size,
  }) async {
    final envelope = await _api.post('/files/presign', data: {
      'file_name': fileName,
      'mime': mime,
      'size': size,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> complete(String fileId, {required int size}) async {
    final envelope = await _api.post('/files/$fileId/complete', data: {'size': size});
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  /// Presign → PUT bytes → complete. Returns completed file DTO.
  Future<Map<String, dynamic>> uploadBytes({
    required String fileName,
    required String mime,
    required List<int> bytes,
  }) async {
    final presign = await this.presign(fileName: fileName, mime: mime, size: bytes.length);
    final uploadUrl = presign['upload_url']?.toString() ?? '';
    if (uploadUrl.isEmpty) {
      throw StateError('missing upload_url');
    }
    await _api.putBytesAbsolute(uploadUrl, bytes, contentType: mime);
    return complete(presign['file_id'].toString(), size: bytes.length);
  }
}
