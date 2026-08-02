import 'dart:convert';

import 'package:drift/drift.dart';
import 'package:sfa_app/core/offline/blob_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Drift-backed [EntityCache] (`cached_entities` + `sync_meta`).
class DriftEntityCache implements EntityCache {
  DriftEntityCache(
    this._db, {
    BlobEntityCache? migrateFrom,
    bool enableBlobMigration = true,
  }) : _migrateFrom = enableBlobMigration ? (migrateFrom ?? BlobEntityCache()) : null;

  static const _metaCursorKey = 'pull_cursor';

  final SfaDatabase _db;
  final BlobEntityCache? _migrateFrom;
  Future<void>? _ready;

  Future<void> _ensureReady() => _ready ??= _migrateBlobIfNeeded();

  Future<void> _migrateBlobIfNeeded() async {
    final blob = _migrateFrom;
    if (blob == null) return;
    final count = await _db.cachedEntities.count().getSingle();
    final existing = await (_db.select(_db.syncMeta)
          ..where((t) => t.key.equals(_metaCursorKey)))
        .getSingleOrNull();
    if (count > 0 || (existing != null && existing.value.isNotEmpty)) {
      return;
    }
    final all = await blob.loadAll();
    if (all.isEmpty) {
      final cursor = await blob.cursor();
      if (cursor != null && cursor.isNotEmpty) {
        await _writeMeta(_metaCursorKey, cursor);
        await blob.clearAfterMigration();
      }
      return;
    }
    final now = DateTime.now().millisecondsSinceEpoch;
    await _db.batch((b) {
      for (final entry in all.entries) {
        for (final entity in entry.value) {
          final id = entity['id']?.toString();
          if (id == null || id.isEmpty) continue;
          b.insert(
            _db.cachedEntities,
            CachedEntitiesCompanion.insert(
              entityType: entry.key,
              entityId: id,
              payloadJson: jsonEncode(entity),
              updatedAt: now,
            ),
            mode: InsertMode.insertOrReplace,
          );
        }
      }
    });
    final cursor = await blob.cursor();
    if (cursor != null && cursor.isNotEmpty) {
      await _writeMeta(_metaCursorKey, cursor);
    }
    await blob.clearAfterMigration();
  }

  Future<void> _writeMeta(String key, String value) {
    return _db.into(_db.syncMeta).insertOnConflictUpdate(
          SyncMetaCompanion.insert(key: key, value: value),
        );
  }

  @override
  Future<void> upsertEntity(String type, Map<String, dynamic> entity) async {
    final id = entity['id']?.toString();
    if (id == null || id.isEmpty) return;
    await _ensureReady();
    await _db.into(_db.cachedEntities).insertOnConflictUpdate(
          CachedEntitiesCompanion.insert(
            entityType: type,
            entityId: id,
            payloadJson: jsonEncode(entity),
            updatedAt: DateTime.now().millisecondsSinceEpoch,
          ),
        );
  }

  @override
  Future<void> deleteEntity(String type, String id) async {
    if (id.isEmpty) return;
    await _ensureReady();
    await (_db.delete(_db.cachedEntities)
          ..where((t) => t.entityType.equals(type) & t.entityId.equals(id)))
        .go();
  }

  @override
  Future<List<Map<String, dynamic>>> listEntities(String type) async {
    await _ensureReady();
    final rows = await (_db.select(_db.cachedEntities)
          ..where((t) => t.entityType.equals(type))
          ..orderBy([(t) => OrderingTerm.asc(t.updatedAt)]))
        .get();
    return rows
        .map((r) => Map<String, dynamic>.from(jsonDecode(r.payloadJson) as Map))
        .toList();
  }

  @override
  Future<String?> cursor() async {
    await _ensureReady();
    final row = await (_db.select(_db.syncMeta)
          ..where((t) => t.key.equals(_metaCursorKey)))
        .getSingleOrNull();
    return row?.value;
  }

  @override
  Future<void> setCursor(String value) async {
    await _ensureReady();
    await _writeMeta(_metaCursorKey, value);
  }
}
