import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/core/offline/sync_worker.dart';

final agentWriteCoordinatorProvider = Provider<AgentWriteCoordinator>((ref) {
  return AgentWriteCoordinator(
    ref.watch(offlineStoreProvider),
    ref.watch(localOutboxProvider),
    ref.watch(syncWorkerProvider),
  );
});

/// Online-first domain write with local cache + outbox fallback.
class AgentWriteCoordinator {
  AgentWriteCoordinator(this._store, this._outbox, this._worker);

  final OfflineStore _store;
  final LocalOutbox _outbox;
  final SyncWorker _worker;

  Future<Map<String, dynamic>> write({
    required String entityType,
    required String op,
    required Map<String, dynamic> payload,
    required Future<Map<String, dynamic>> Function() online,
    int baseVersion = 0,
  }) async {
    final opId = 'local-${DateTime.now().microsecondsSinceEpoch}';
    try {
      final created = await online();
      final id = created['id']?.toString() ?? payload['id']?.toString() ?? opId;
      final entity = {...created, 'id': id};
      await _store.upsertEntity(entityType, entity);
      // Best-effort immediate sync so peers pull RecordChange fan-out.
      await _worker.tick(reason: 'agent-write');
      return entity;
    } catch (_) {
      final id = payload['id']?.toString() ?? opId;
      final entity = {...payload, 'id': id};
      await _store.upsertEntity(entityType, entity);
      await _outbox.enqueue(OutboxOp(
        opId: opId,
        entityType: entityType,
        entityId: id,
        op: op,
        baseVersion: baseVersion,
        payload: entity,
      ));
      return entity;
    }
  }
}
