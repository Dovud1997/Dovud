import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/core/offline/sqlite_entity_cache.dart';
import 'package:sqflite_common_ffi/sqflite_ffi.dart';

void main() {
  setUpAll(() {
    sqfliteFfiInit();
  });

  test('SqliteEntityCache upsert list delete cursor', () async {
    final cache = SqliteEntityCache(
      databaseFactory: databaseFactoryFfi,
      dbPath: inMemoryDatabasePath,
      enableBlobMigration: false,
    );

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

    await cache.close();
  });
}
