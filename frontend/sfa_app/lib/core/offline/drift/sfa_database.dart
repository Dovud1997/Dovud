import 'dart:io';

import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:sqlite3_flutter_libs/sqlite3_flutter_libs.dart';

part 'sfa_database.g.dart';

@DataClassName('CachedEntityRow')
class CachedEntities extends Table {
  TextColumn get entityType => text()();
  TextColumn get entityId => text()();
  TextColumn get payloadJson => text()();
  IntColumn get updatedAt => integer()();

  @override
  Set<Column<Object>> get primaryKey => {entityType, entityId};
}

class SyncMeta extends Table {
  TextColumn get key => text()();
  TextColumn get value => text()();

  @override
  Set<Column<Object>> get primaryKey => {key};
}

@DataClassName('OutboxOpRow')
class OutboxOps extends Table {
  TextColumn get opId => text()();
  TextColumn get entityType => text()();
  TextColumn get entityId => text()();
  TextColumn get op => text()();
  IntColumn get baseVersion => integer().withDefault(const Constant(0))();
  TextColumn get payloadJson => text()();
  TextColumn get status => text().withDefault(const Constant('pending'))();
  DateTimeColumn get createdAt => dateTime()();

  @override
  Set<Column<Object>> get primaryKey => {opId};
}

@DriftDatabase(tables: [CachedEntities, SyncMeta, OutboxOps])
class SfaDatabase extends _$SfaDatabase {
  SfaDatabase(super.e);

  /// In-memory DB for unit tests.
  SfaDatabase.memory() : super(NativeDatabase.memory());

  /// Persistent app DB under documents directory.
  factory SfaDatabase.open() {
    return SfaDatabase(LazyDatabase(() async {
      if (Platform.isAndroid) {
        await applyWorkaroundToOpenSqlite3OnOldAndroidVersions();
      }
      final dir = await getApplicationDocumentsDirectory();
      final file = File(p.join(dir.path, 'sfa_offline_v1.db'));
      return NativeDatabase.createInBackground(file);
    }));
  }

  @override
  int get schemaVersion => 1;
}
