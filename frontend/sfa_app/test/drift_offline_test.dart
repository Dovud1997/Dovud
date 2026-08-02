import 'dart:typed_data';

import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/drift_outbox_store.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/file_upload_queue.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:drift/drift.dart' show Value;

void main() {
  late SfaDatabase db;

  setUp(() {
    db = SfaDatabase.memory(NativeDatabase.memory());
  });

  tearDown(() async {
    await db.close();
  });

  test('DriftEntityCache upsert list delete cursor', () async {
    final cache = DriftEntityCache(db, enableBlobMigration: false);

    await cache.upsertEntity('product', {'id': 'p1', 'name': 'A'});
    await cache.upsertEntity('product', {'id': 'p2', 'name': 'B'});
    await cache.upsertEntity('product', {'id': 'p1', 'name': 'A2'});
    await cache.setCursor('cur-1');

    final list = await cache.listEntities('product');
    expect(list.length, 2);
    expect(list.firstWhere((e) => e['id'] == 'p1')['name'], 'A2');
    expect(await cache.cursor(), 'cur-1');

    await cache.deleteEntity('product', 'p2');
    expect((await cache.listEntities('product')).length, 1);
  });

  test('DriftOutboxStore enqueue markStatus remove', () async {
    final store = DriftOutboxStore(db, enableBlobMigration: false);

    await store.enqueue(OutboxOp(
      opId: 'op-1',
      entityType: 'note',
      entityId: 'n1',
      op: 'create',
      payload: {'text': 'hi'},
    ));
    await store.enqueue(OutboxOp(
      opId: 'op-2',
      entityType: 'note',
      entityId: 'n2',
      op: 'create',
    ));

    expect((await store.list()).length, 2);
    await store.markStatus('op-1', 'conflict');
    expect((await store.list(status: 'pending')).length, 1);
    expect((await store.list(status: 'conflict')).single.opId, 'op-1');
    await store.removeByOpIds(['op-1']);
    expect((await store.list(status: 'conflict')), isEmpty);
  });

  test('FileUploads table enqueue', () async {
    await db.into(db.fileUploads).insert(
          FileUploadsCompanion.insert(
            uploadId: 'up-1',
            fileName: 'a.txt',
            mime: 'text/plain',
            sizeBytes: const Value(3),
            status: const Value('pending'),
            createdAt: DateTime.now().toUtc(),
            payload: Value(Uint8List.fromList([1, 2, 3])),
          ),
        );
    final rows = await db.select(db.fileUploads).get();
    expect(rows.length, 1);
    expect(rows.first.fileName, 'a.txt');
    expect(rows.first.payload, [1, 2, 3]);
  });

  test('FileUploadQueue survives without in-memory cache', () async {
    final uploaded = <List<int>>[];
    final queue = FileUploadQueue(
      db,
      uploadBytes: ({required fileName, required mime, required bytes}) async {
        uploaded.add(List<int>.from(bytes));
        return {'id': 'file-1', 'file_name': fileName};
      },
    );
    final pending = await queue.enqueue(
      fileName: 'note.bin',
      mime: 'application/octet-stream',
      bytes: [9, 8, 7],
    );
    // Simulate process restart: new queue instance, empty memory map.
    final restored = FileUploadQueue(
      db,
      uploadBytes: ({required fileName, required mime, required bytes}) async {
        uploaded.add(List<int>.from(bytes));
        return {'id': 'file-1', 'file_name': fileName};
      },
    );
    final listed = await restored.list(status: 'pending');
    expect(listed.single.uploadId, pending.uploadId);
    expect(listed.single.bytes, [9, 8, 7]);

    final flush = await restored.flush();
    expect(flush['uploaded'], 1);
    expect(flush['failed'], 0);
    expect(uploaded.length, 1);
    expect(uploaded.first, [9, 8, 7]);

    final row = await (db.select(db.fileUploads)
          ..where((t) => t.uploadId.equals(pending.uploadId)))
        .getSingle();
    expect(row.status, 'uploaded');
    expect(row.payload, isNull);
  });
}
