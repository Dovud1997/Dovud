import 'package:sfa_app/core/offline/local_outbox.dart';

/// Persistence for the local sync outbox (Drift table or encrypted blob).
abstract class OutboxStore {
  Future<List<OutboxOp>> list({String? status});

  Future<void> enqueue(OutboxOp op);

  Future<void> clear();

  Future<void> removeByOpIds(Iterable<String> ids);

  Future<void> markStatus(String opId, String status);
}
