import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/drift_outbox_store.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';

void main() {
  late SfaDatabase db;

  setUp(() {
    db = SfaDatabase.memory();
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

  test('DriftOutboxStore enqueue list remove', () async {
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
    await store.removeByOpIds(['op-1']);
    final left = await store.list();
    expect(left.length, 1);
    expect(left.first.opId, 'op-2');
  });
}
