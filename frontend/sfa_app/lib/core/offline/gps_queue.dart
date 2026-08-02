import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/features/fieldforce/data/fieldforce_repository.dart';

final gpsQueueProvider = Provider<GpsQueue>((ref) {
  return GpsQueue(
    sharedSfaDatabase(),
    ref.watch(fieldForceRepositoryProvider),
  );
});

class PendingGpsPoint {
  PendingGpsPoint({
    required this.pointId,
    required this.agentId,
    this.visitId,
    required this.lat,
    required this.lng,
    this.accuracy,
    required this.recordedAt,
    required this.status,
    this.error,
  });

  final String pointId;
  final String agentId;
  final String? visitId;
  final double lat;
  final double lng;
  final double? accuracy;
  final DateTime recordedAt;
  final String status;
  final String? error;
}

/// Offline GPS batch queue backed by Drift `gps_pending`.
class GpsQueue {
  GpsQueue(this._db, this._ff);

  final SfaDatabase _db;
  final FieldForceRepository _ff;

  Future<PendingGpsPoint> enqueue({
    required String agentId,
    required double lat,
    required double lng,
    String? visitId,
    double? accuracy,
    DateTime? recordedAt,
  }) async {
    final id = 'gps-${DateTime.now().microsecondsSinceEpoch}';
    final at = (recordedAt ?? DateTime.now().toUtc()).toUtc();
    await _db.into(_db.gpsPending).insert(
          GpsPendingCompanion.insert(
            pointId: id,
            agentId: agentId,
            visitId: Value(visitId),
            lat: lat,
            lng: lng,
            accuracy: Value(accuracy),
            recordedAt: at,
            status: const Value('pending'),
            createdAt: DateTime.now().toUtc(),
          ),
        );
    return PendingGpsPoint(
      pointId: id,
      agentId: agentId,
      visitId: visitId,
      lat: lat,
      lng: lng,
      accuracy: accuracy,
      recordedAt: at,
      status: 'pending',
    );
  }

  Future<List<PendingGpsPoint>> list({String status = 'pending'}) async {
    final rows = await (_db.select(_db.gpsPending)
          ..where((t) => t.status.equals(status))
          ..orderBy([(t) => OrderingTerm.asc(t.recordedAt)]))
        .get();
    return rows
        .map(
          (r) => PendingGpsPoint(
            pointId: r.pointId,
            agentId: r.agentId,
            visitId: r.visitId,
            lat: r.lat,
            lng: r.lng,
            accuracy: r.accuracy,
            recordedAt: r.recordedAt.toUtc(),
            status: r.status,
            error: r.error,
          ),
        )
        .toList();
  }

  Future<int> pendingCount() async => (await list()).length;

  Future<Map<String, dynamic>> flush() async {
    final pending = await list(status: 'pending');
    if (pending.isEmpty) {
      return {'pending': 0, 'uploaded': 0, 'failed': 0};
    }
    final payload = pending
        .map(
          (p) => {
            'agent_id': p.agentId,
            if (p.visitId != null && p.visitId!.isNotEmpty) 'visit_id': p.visitId,
            'lat': p.lat,
            'lng': p.lng,
            if (p.accuracy != null) 'accuracy': p.accuracy,
            'recorded_at': p.recordedAt.toUtc().toIso8601String(),
          },
        )
        .toList();
    try {
      await _ff.uploadGpsPoints(payload);
      for (final p in pending) {
        await (_db.update(_db.gpsPending)..where((t) => t.pointId.equals(p.pointId)))
            .write(const GpsPendingCompanion(status: Value('uploaded')));
      }
      return {'pending': pending.length, 'uploaded': pending.length, 'failed': 0};
    } catch (e) {
      return {
        'pending': pending.length,
        'uploaded': 0,
        'failed': pending.length,
        'error': e.toString(),
      };
    }
  }
}
