import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';
import 'package:sfa_app/core/offline/entity_cache_factory.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final offlineStoreProvider = Provider<OfflineStore>((ref) {
  return OfflineStore(
    ref.watch(syncRepositoryProvider),
    ref.watch(localOutboxProvider),
    cache: createEntityCache(),
  );
});

/// Sync orchestration + [EntityCache] facade.
class OfflineStore implements EntityCache {
  OfflineStore(this._sync, this.outbox, {EntityCache? cache})
      : _cache = cache ?? createEntityCache();

  final SyncRepository _sync;
  final LocalOutbox outbox;
  final EntityCache _cache;

  String get cacheBackendLabel => entityCacheBackendLabel();

  @override
  Future<void> upsertEntity(String type, Map<String, dynamic> entity) =>
      _cache.upsertEntity(type, entity);

  @override
  Future<void> deleteEntity(String type, String id) =>
      _cache.deleteEntity(type, id);

  @override
  Future<List<Map<String, dynamic>>> listEntities(String type) =>
      _cache.listEntities(type);

  @override
  Future<String?> cursor() => _cache.cursor();

  @override
  Future<void> setCursor(String value) => _cache.setCursor(value);

  Future<Map<String, dynamic>> pullAndCache({String? deviceId}) async {
    final cur = await cursor() ?? '';
    final res = await _sync.pull(deviceId: deviceId, cursor: cur);
    final changes = (res['changes'] as List?) ?? const [];
    final types = <String>{};
    for (final c in changes) {
      final m = Map<String, dynamic>.from(c as Map);
      final type = m['entity_type']?.toString() ?? 'unknown';
      types.add(type);
      final payload = Map<String, dynamic>.from(
        (m['data'] as Map?) ?? (m['payload'] as Map?) ?? const {},
      );
      if (payload['id'] == null && m['entity_id'] != null) {
        payload['id'] = m['entity_id'];
      }
      if (m['deleted'] == true) {
        await deleteEntity(type, m['entity_id']?.toString() ?? '');
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
      'cached_types': types.toList(),
      'backend': cacheBackendLabel,
    };
  }

  Future<Map<String, dynamic>> flushOutbox({String? deviceId}) {
    return outbox.flush(deviceId: deviceId);
  }

  /// Push pending outbox then pull & cache (background sync cycle).
  Future<Map<String, dynamic>> syncCycle({String? deviceId}) async {
    final flush = await flushOutbox(deviceId: deviceId);
    final pull = await pullAndCache(deviceId: deviceId);
    return {
      'flush': flush,
      'pull': pull,
      'backend': cacheBackendLabel,
    };
  }
}
