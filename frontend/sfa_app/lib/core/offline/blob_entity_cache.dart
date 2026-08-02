import 'dart:convert';

import 'package:sfa_app/core/offline/entity_cache.dart';
import 'package:sfa_app/core/offline/secure_blob_store.dart';

/// Encrypted SharedPreferences-backed [EntityCache] (web + migration source).
class BlobEntityCache implements EntityCache {
  BlobEntityCache({SecureBlobStore? blobs}) : _blobs = blobs ?? SecureBlobStore();

  static const entitiesKey = 'sfa_offline_entities_v1';
  static const cursorKey = 'sfa_offline_cursor_v1';

  final SecureBlobStore _blobs;

  Future<Map<String, List<Map<String, dynamic>>>> loadAll() async {
    final raw = await _blobs.read(entitiesKey);
    if (raw == null || raw.isEmpty) return {};
    final decoded = jsonDecode(raw) as Map<String, dynamic>;
    return decoded.map((k, v) {
      final list = (v as List?) ?? const [];
      return MapEntry(
        k,
        list.map((e) => Map<String, dynamic>.from(e as Map)).toList(),
      );
    });
  }

  Future<void> _saveAll(Map<String, List<Map<String, dynamic>>> data) async {
    await _blobs.write(entitiesKey, jsonEncode(data));
  }

  @override
  Future<void> upsertEntity(String type, Map<String, dynamic> entity) async {
    final id = entity['id']?.toString();
    if (id == null || id.isEmpty) return;
    final all = await loadAll();
    final list = all[type] ?? <Map<String, dynamic>>[];
    final idx = list.indexWhere((e) => e['id']?.toString() == id);
    if (idx >= 0) {
      list[idx] = entity;
    } else {
      list.add(entity);
    }
    all[type] = list;
    await _saveAll(all);
  }

  @override
  Future<void> deleteEntity(String type, String id) async {
    if (id.isEmpty) return;
    final all = await loadAll();
    final list = all[type] ?? <Map<String, dynamic>>[];
    all[type] = list.where((e) => e['id']?.toString() != id).toList();
    await _saveAll(all);
  }

  @override
  Future<List<Map<String, dynamic>>> listEntities(String type) async {
    final all = await loadAll();
    return all[type] ?? const [];
  }

  @override
  Future<String?> cursor() => _blobs.read(cursorKey);

  @override
  Future<void> setCursor(String value) => _blobs.write(cursorKey, value);

  /// Clears blob keys after a successful migrate-to-SQLite.
  Future<void> clearAfterMigration() async {
    await _blobs.remove(entitiesKey);
    await _blobs.remove(cursorKey);
  }
}
