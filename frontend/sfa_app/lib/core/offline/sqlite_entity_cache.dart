import 'dart:convert';

import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:sfa_app/core/offline/blob_entity_cache.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';
import 'package:sqflite/sqflite.dart';

/// SQLite-backed [EntityCache] (Drift-shaped tables; mobile / desktop).
///
/// Schema mirrors architecture `local_entities` + `sync_cursors`:
/// - `cached_entities(entity_type, entity_id, payload_json, updated_at)`
/// - `sync_meta(key, value)` for the pull cursor
class SqliteEntityCache implements EntityCache {
  SqliteEntityCache({
    DatabaseFactory? databaseFactory,
    String? dbPath,
    BlobEntityCache? migrateFrom,
    bool enableBlobMigration = true,
  })  : _databaseFactory = databaseFactory,
        _dbPath = dbPath,
        _migrateFrom = enableBlobMigration
            ? (migrateFrom ?? BlobEntityCache())
            : null;

  static const _dbName = 'sfa_entity_cache_v1.db';
  static const _metaCursorKey = 'pull_cursor';

  final DatabaseFactory? _databaseFactory;
  final String? _dbPath;
  final BlobEntityCache? _migrateFrom;

  Database? _db;
  Future<Database>? _opening;

  Future<Database> _ensureDb() {
    if (_db != null) return Future.value(_db);
    return _opening ??= _open();
  }

  Future<Database> _open() async {
    final factory = _databaseFactory ?? databaseFactory;
    final path = _dbPath ??
        p.join((await getApplicationDocumentsDirectory()).path, _dbName);
    final db = await factory.openDatabase(
      path,
      options: OpenDatabaseOptions(
        version: 1,
        onCreate: (db, version) async {
          await db.execute('''
CREATE TABLE cached_entities (
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (entity_type, entity_id)
)''');
          await db.execute('''
CREATE TABLE sync_meta (
  key TEXT PRIMARY KEY NOT NULL,
  value TEXT NOT NULL
)''');
        },
      ),
    );
    _db = db;
    await _migrateBlobIfNeeded(db);
    return db;
  }

  Future<void> _migrateBlobIfNeeded(Database db) async {
    final blob = _migrateFrom;
    if (blob == null) return;
    final count = Sqflite.firstIntValue(
          await db.rawQuery('SELECT COUNT(*) AS c FROM cached_entities'),
        ) ??
        0;
    final existingCursor = await _readMeta(db, _metaCursorKey);
    if (count > 0 || (existingCursor != null && existingCursor.isNotEmpty)) {
      return;
    }
    final all = await blob.loadAll();
    if (all.isEmpty) {
      final blobCursor = await blob.cursor();
      if (blobCursor != null && blobCursor.isNotEmpty) {
        await _writeMeta(db, _metaCursorKey, blobCursor);
        await blob.clearAfterMigration();
      }
      return;
    }
    final batch = db.batch();
    final now = DateTime.now().millisecondsSinceEpoch;
    for (final entry in all.entries) {
      for (final entity in entry.value) {
        final id = entity['id']?.toString();
        if (id == null || id.isEmpty) continue;
        batch.insert(
          'cached_entities',
          {
            'entity_type': entry.key,
            'entity_id': id,
            'payload_json': jsonEncode(entity),
            'updated_at': now,
          },
          conflictAlgorithm: ConflictAlgorithm.replace,
        );
      }
    }
    await batch.commit(noResult: true);
    final blobCursor = await blob.cursor();
    if (blobCursor != null && blobCursor.isNotEmpty) {
      await _writeMeta(db, _metaCursorKey, blobCursor);
    }
    await blob.clearAfterMigration();
  }

  Future<String?> _readMeta(Database db, String key) async {
    final rows = await db.query(
      'sync_meta',
      columns: ['value'],
      where: 'key = ?',
      whereArgs: [key],
      limit: 1,
    );
    if (rows.isEmpty) return null;
    return rows.first['value'] as String?;
  }

  Future<void> _writeMeta(Database db, String key, String value) async {
    await db.insert(
      'sync_meta',
      {'key': key, 'value': value},
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  @override
  Future<void> upsertEntity(String type, Map<String, dynamic> entity) async {
    final id = entity['id']?.toString();
    if (id == null || id.isEmpty) return;
    final db = await _ensureDb();
    await db.insert(
      'cached_entities',
      {
        'entity_type': type,
        'entity_id': id,
        'payload_json': jsonEncode(entity),
        'updated_at': DateTime.now().millisecondsSinceEpoch,
      },
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }

  @override
  Future<void> deleteEntity(String type, String id) async {
    if (id.isEmpty) return;
    final db = await _ensureDb();
    await db.delete(
      'cached_entities',
      where: 'entity_type = ? AND entity_id = ?',
      whereArgs: [type, id],
    );
  }

  @override
  Future<List<Map<String, dynamic>>> listEntities(String type) async {
    final db = await _ensureDb();
    final rows = await db.query(
      'cached_entities',
      columns: ['payload_json'],
      where: 'entity_type = ?',
      whereArgs: [type],
      orderBy: 'updated_at ASC',
    );
    return rows.map((r) {
      final raw = r['payload_json'] as String? ?? '{}';
      return Map<String, dynamic>.from(jsonDecode(raw) as Map);
    }).toList();
  }

  @override
  Future<String?> cursor() async {
    final db = await _ensureDb();
    return _readMeta(db, _metaCursorKey);
  }

  @override
  Future<void> setCursor(String value) async {
    final db = await _ensureDb();
    await _writeMeta(db, _metaCursorKey, value);
  }

  /// Storage backend label for Sync UI.
  String get backendLabel => 'SQLite tables';

  Future<void> close() async {
    await _db?.close();
    _db = null;
    _opening = null;
  }
}
