import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

final localOutboxProvider = Provider<LocalOutbox>((ref) {
  return LocalOutbox(ref.watch(syncRepositoryProvider));
});

class OutboxOp {
  OutboxOp({
    required this.opId,
    required this.entityType,
    required this.entityId,
    required this.op,
    this.baseVersion = 0,
    this.payload = const {},
    DateTime? createdAt,
  }) : createdAt = createdAt ?? DateTime.now().toUtc();

  final String opId;
  final String entityType;
  final String entityId;
  final String op;
  final int baseVersion;
  final Map<String, dynamic> payload;
  final DateTime createdAt;

  Map<String, dynamic> toJson() => {
        'op_id': opId,
        'entity_type': entityType,
        'entity_id': entityId,
        'op': op,
        'base_version': baseVersion,
        'payload': payload,
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

class LocalOutbox {
  LocalOutbox(this._sync);

  static const _key = 'sfa_local_outbox_v1';
  final SyncRepository _sync;

  Future<List<OutboxOp>> list() async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(_key);
    if (raw == null || raw.isEmpty) return const [];
    final list = jsonDecode(raw) as List<dynamic>;
    return list.map((e) => OutboxOp.fromJson(Map<String, dynamic>.from(e as Map))).toList();
  }

  Future<void> enqueue(OutboxOp op) async {
    final prefs = await SharedPreferences.getInstance();
    final current = await list();
    current.add(op);
    await prefs.setString(_key, jsonEncode(current.map((e) => e.toJson()).toList()));
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_key);
  }

  Future<void> removeByOpIds(Iterable<String> ids) async {
    final set = ids.toSet();
    final remaining = (await list()).where((o) => !set.contains(o.opId)).toList();
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_key, jsonEncode(remaining.map((e) => e.toJson()).toList()));
  }

  /// Push pending local ops to `/sync/push` and drop acknowledged ones.
  Future<Map<String, dynamic>> flush({String deviceId = 'flutter-web'}) async {
    final pending = await list();
    if (pending.isEmpty) {
      return {'pushed': 0, 'acked': 0, 'conflicts': 0, 'rejected': 0};
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
      } else {
        rejected++;
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
      'remaining': (await list()).length,
    };
  }
}
