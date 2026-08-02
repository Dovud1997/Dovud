import 'dart:convert';

import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';
import 'package:sfa_app/core/offline/secure_blob_store.dart';

/// Encrypted blob outbox (migration source / emergency fallback).
class BlobOutboxStore implements OutboxStore {
  BlobOutboxStore({SecureBlobStore? blobs}) : _blobs = blobs ?? SecureBlobStore();

  static const key = 'sfa_local_outbox_v1';

  final SecureBlobStore _blobs;

  Future<List<OutboxOp>> _all() async {
    final raw = await _blobs.read(key);
    if (raw == null || raw.isEmpty) return const [];
    final list = jsonDecode(raw) as List<dynamic>;
    return list.map((e) => OutboxOp.fromJson(Map<String, dynamic>.from(e as Map))).toList();
  }

  Future<void> _save(List<OutboxOp> ops) async {
    await _blobs.write(key, jsonEncode(ops.map((e) => e.toJson()).toList()));
  }

  @override
  Future<List<OutboxOp>> list({String? status}) async {
    final all = await _all();
    final filter = (status == null || status.isEmpty) ? 'pending' : status;
    return all.where((o) => o.status == filter).toList();
  }

  @override
  Future<void> enqueue(OutboxOp op) async {
    final current = await _all();
    op.status = 'pending';
    current.add(op);
    await _save(current);
  }

  @override
  Future<void> clear() => _blobs.remove(key);

  @override
  Future<void> removeByOpIds(Iterable<String> ids) async {
    final set = ids.toSet();
    final remaining = (await _all()).where((o) => !set.contains(o.opId)).toList();
    await _save(remaining);
  }

  @override
  Future<void> markStatus(String opId, String status) async {
    final all = await _all();
    for (final o in all) {
      if (o.opId == opId) o.status = status;
    }
    await _save(all);
  }
}
