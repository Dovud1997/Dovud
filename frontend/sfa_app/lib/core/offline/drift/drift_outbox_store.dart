import 'dart:convert';

import 'package:drift/drift.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';
import 'package:sfa_app/core/offline/secure_blob_store.dart';

/// Drift-backed outbox (`outbox_ops` table).
class DriftOutboxStore implements OutboxStore {
  DriftOutboxStore(
    this._db, {
    SecureBlobStore? migrateFrom,
    bool enableBlobMigration = true,
  }) : _migrateFrom = enableBlobMigration ? (migrateFrom ?? SecureBlobStore()) : null;

  static const _blobKey = 'sfa_local_outbox_v1';

  final SfaDatabase _db;
  final SecureBlobStore? _migrateFrom;
  Future<void>? _ready;

  Future<void> _ensureReady() => _ready ??= _migrateBlobIfNeeded();

  Future<void> _migrateBlobIfNeeded() async {
    final blobs = _migrateFrom;
    if (blobs == null) return;
    final count = await _db.outboxOps.count().getSingle();
    if (count > 0) return;
    final raw = await blobs.read(_blobKey);
    if (raw == null || raw.isEmpty) return;
    final list = jsonDecode(raw) as List<dynamic>;
    await _db.batch((b) {
      for (final e in list) {
        final op = OutboxOp.fromJson(Map<String, dynamic>.from(e as Map));
        if (op.opId.isEmpty) continue;
        b.insert(
          _db.outboxOps,
          OutboxOpsCompanion.insert(
            opId: op.opId,
            entityType: op.entityType,
            entityId: op.entityId,
            op: op.op,
            baseVersion: Value(op.baseVersion),
            payloadJson: jsonEncode(op.payload),
            status: Value(op.status),
            createdAt: op.createdAt,
          ),
          mode: InsertMode.insertOrReplace,
        );
      }
    });
    await blobs.remove(_blobKey);
  }

  OutboxOp _toOp(OutboxOpRow r) => OutboxOp(
        opId: r.opId,
        entityType: r.entityType,
        entityId: r.entityId,
        op: r.op,
        baseVersion: r.baseVersion,
        payload: Map<String, dynamic>.from(jsonDecode(r.payloadJson) as Map),
        status: r.status,
        createdAt: r.createdAt.toUtc(),
      );

  @override
  Future<List<OutboxOp>> list({String? status}) async {
    await _ensureReady();
    final filter = (status == null || status.isEmpty) ? 'pending' : status;
    final rows = await (_db.select(_db.outboxOps)
          ..where((t) => t.status.equals(filter))
          ..orderBy([(t) => OrderingTerm.asc(t.createdAt)]))
        .get();
    return rows.map(_toOp).toList();
  }

  @override
  Future<void> enqueue(OutboxOp op) async {
    await _ensureReady();
    await _db.into(_db.outboxOps).insertOnConflictUpdate(
          OutboxOpsCompanion.insert(
            opId: op.opId,
            entityType: op.entityType,
            entityId: op.entityId,
            op: op.op,
            baseVersion: Value(op.baseVersion),
            payloadJson: jsonEncode(op.payload),
            status: const Value('pending'),
            createdAt: op.createdAt,
          ),
        );
  }

  @override
  Future<void> clear() async {
    await _ensureReady();
    await _db.delete(_db.outboxOps).go();
  }

  @override
  Future<void> removeByOpIds(Iterable<String> ids) async {
    final set = ids.toSet();
    if (set.isEmpty) return;
    await _ensureReady();
    await (_db.delete(_db.outboxOps)..where((t) => t.opId.isIn(set))).go();
  }

  @override
  Future<void> markStatus(String opId, String status) async {
    if (opId.isEmpty) return;
    await _ensureReady();
    await (_db.update(_db.outboxOps)..where((t) => t.opId.equals(opId))).write(
          OutboxOpsCompanion(status: Value(status)),
        );
  }
}
