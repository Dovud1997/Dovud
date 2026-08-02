import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/secure_blob_store.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final offlineStoreProvider = Provider<OfflineStore>((ref) {
  return OfflineStore(ref.watch(syncRepositoryProvider), ref.watch(localOutboxProvider));
});

/// Encrypted local cache (SharedPreferences + secure key). Swap to Drift/Isar later.
class OfflineStore {
  OfflineStore(this._sync, this.outbox, {SecureBlobStore? blobs})
      : _blobs = blobs ?? SecureBlobStore();

  static const _entitiesKey = 'sfa_offline_entities_v1';
  static const _cursorKey = 'sfa_offline_cursor_v1';

  final SyncRepository _sync;
  final LocalOutbox outbox;
  final SecureBlobStore _blobs;

  Future<Map<String, List<Map<String, dynamic>>>> _loadEntities() async {
    final raw = await _blobs.read(_entitiesKey);
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

  Future<void> _saveEntities(Map<String, List<Map<String, dynamic>>> data) async {
    await _blobs.write(_entitiesKey, jsonEncode(data));
  }

  Future<void> upsertEntity(String type, Map<String, dynamic> entity) async {
    final id = entity['id']?.toString();
    if (id == null || id.isEmpty) return;
    final all = await _loadEntities();
    final list = all[type] ?? <Map<String, dynamic>>[];
    final idx = list.indexWhere((e) => e['id']?.toString() == id);
    if (idx >= 0) {
      list[idx] = entity;
    } else {
      list.add(entity);
    }
    all[type] = list;
    await _saveEntities(all);
  }

  Future<List<Map<String, dynamic>>> listEntities(String type) async {
    final all = await _loadEntities();
    return all[type] ?? const [];
  }

  Future<String?> cursor() async {
    return _blobs.read(_cursorKey);
  }

  Future<void> setCursor(String value) async {
    await _blobs.write(_cursorKey, value);
  }

  Future<Map<String, dynamic>> pullAndCache({String deviceId = 'flutter-web'}) async {
    final cur = await cursor() ?? '';
    final res = await _sync.pull(deviceId: deviceId, cursor: cur);
    final changes = (res['changes'] as List?) ?? const [];
    for (final c in changes) {
      final m = Map<String, dynamic>.from(c as Map);
      final type = m['entity_type']?.toString() ?? 'unknown';
      final payload = Map<String, dynamic>.from(m['payload'] as Map? ?? const {});
      if (payload['id'] == null && m['entity_id'] != null) {
        payload['id'] = m['entity_id'];
      }
      if (m['deleted'] == true) {
        final all = await _loadEntities();
        final list = all[type] ?? <Map<String, dynamic>>[];
        all[type] = list.where((e) => e['id']?.toString() != m['entity_id']?.toString()).toList();
        await _saveEntities(all);
      } else {
        await upsertEntity(type, payload);
      }
    }
    final next = res['next_cursor']?.toString() ?? res['cursor']?.toString() ?? '';
    if (next.isNotEmpty) {
      await setCursor(next);
    }
    return {
      'changes': changes.length,
      'cursor': next,
      'cached_types': (await _loadEntities()).keys.toList(),
    };
  }

  Future<Map<String, dynamic>> flushOutbox({String deviceId = 'flutter-web'}) {
    return outbox.flush(deviceId: deviceId);
  }
}
