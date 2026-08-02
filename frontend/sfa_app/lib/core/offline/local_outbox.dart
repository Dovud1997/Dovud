import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/outbox_factory.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final localOutboxProvider = Provider<LocalOutbox>((ref) {
  return LocalOutbox(
    ref.watch(syncRepositoryProvider),
    store: createOutboxStore(),
  );
});

class OutboxOp {
  OutboxOp({
    required this.opId,
    required this.entityType,
    required this.entityId,
    required this.op,
    this.baseVersion = 0,
    this.payload = const {},
    this.status = 'pending',
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now().toUtc();

  final String opId;
  final String entityType;
  final String entityId;
  final String op;
  final int baseVersion;
  final Map<String, dynamic> payload;
  String status;
  final DateTime createdAt;

  Map<String, dynamic> toJson() => {
        'op_id': opId,
        'entity_type': entityType,
        'entity_id': entityId,
        'op': op,
        'base_version': baseVersion,
        'payload': payload,
        'status': status,
        'created_at': createdAt.toIso8601String(),
      };

  factory OutboxOp.fromJson(Map<String, dynamic> json) {
    return OutboxOp(
      opId: json['op_id']?.toString() ?? '',
      entityType: json['entity_type']?.toString() ?? '',
      entityId: json['entity_id']?.toString() ?? '',
      op: json['op']?.toString() ?? 'update',
      baseVersion: (json['base_version'] as num?)?.toInt() ?? 0,
      payload: Map<String, dynamic>.from(json['payload'] as Map? ?? const {}),
      status: json['status']?.toString() ?? 'pending',
      createdAt: DateTime.tryParse(json['created_at']?.toString() ?? '') ?? DateTime.now().toUtc(),
    );
  }

  Map<String, dynamic> toSyncOp() => {
        'op_id': opId,
        'entity_type': entityType,
        'entity_id': entityId,
        'op': op,
        'base_version': baseVersion,
        'payload': payload,
      };
}

/// Sync outbox orchestration; persistence via [OutboxStore] (Drift or blob).
class LocalOutbox {
  LocalOutbox(this._sync, {OutboxStore? store}) : _store = store ?? createOutboxStore();

  final SyncRepository _sync;
  final OutboxStore _store;

  String get backendLabel => outboxBackendLabel();

  Future<List<OutboxOp>> list({String? status}) => _store.list(status: status);

  Future<void> enqueue(OutboxOp op) => _store.enqueue(op);

  Future<void> clear() => _store.clear();

  Future<void> removeByOpIds(Iterable<String> ids) => _store.removeByOpIds(ids);

  Future<void> markStatus(String opId, String status) => _store.markStatus(opId, status);

  Future<Map<String, dynamic>> flush({String? deviceId}) async {
    final pending = await list(status: 'pending');
    if (pending.isEmpty) {
      return {'pushed': 0, 'acked': 0, 'conflicts': 0, 'rejected': 0, 'backend': backendLabel};
    }
    final res = await _sync.push(
      deviceId: deviceId,
      ops: pending.map((e) => e.toSyncOp()).toList(),
    );
    final results = (res['results'] as List?) ?? const [];
    final acked = <String>[];
    var conflicts = 0;
    var rejected = 0;
    for (final r in results) {
      final m = Map<String, dynamic>.from(r as Map);
      final opId = m['op_id']?.toString() ?? '';
      final status = m['status']?.toString() ?? '';
      if (status == 'acked') {
        acked.add(opId);
      } else if (status == 'conflict') {
        conflicts++;
        if (opId.isNotEmpty) await markStatus(opId, 'conflict');
      } else {
        rejected++;
        if (opId.isNotEmpty) await markStatus(opId, 'rejected');
      }
    }
    if (acked.isNotEmpty) {
      await removeByOpIds(acked);
    }
    return {
      'pushed': pending.length,
      'acked': acked.length,
      'conflicts': conflicts,
      'rejected': rejected,
      'remaining': (await list(status: 'pending')).length,
      'backend': backendLabel,
    };
  }
}
