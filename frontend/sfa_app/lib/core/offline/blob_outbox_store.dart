import 'dart:convert';

import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';
import 'package:sfa_app/core/offline/secure_blob_store.dart';

/// Encrypted blob outbox (web + migration source).
class BlobOutboxStore implements OutboxStore {
  BlobOutboxStore({SecureBlobStore? blobs}) : _blobs = blobs ?? SecureBlobStore();

  static const key = 'sfa_local_outbox_v1';

  final SecureBlobStore _blobs;

  @override
  Future<List<OutboxOp>> list() async {
    final raw = await _blobs.read(key);
    if (raw == null || raw.isEmpty) return const [];
    final list = jsonDecode(raw) as List<dynamic>;
    return list.map((e) => OutboxOp.fromJson(Map<String, dynamic>.from(e as Map))).toList();
  }

  @override
  Future<void> enqueue(OutboxOp op) async {
    final current = await list();
    current.add(op);
    await _blobs.write(key, jsonEncode(current.map((e) => e.toJson()).toList()));
  }

  @override
  Future<void> clear() => _blobs.remove(key);

  @override
  Future<void> removeByOpIds(Iterable<String> ids) async {
    final set = ids.toSet();
    final remaining = (await list()).where((o) => !set.contains(o.opId)).toList();
    await _blobs.write(key, jsonEncode(remaining.map((e) => e.toJson()).toList()));
  }
}
