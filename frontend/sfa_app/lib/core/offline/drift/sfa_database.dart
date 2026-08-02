import 'package:drift/drift.dart';
import 'package:drift_flutter/drift_flutter.dart';

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

@DataClassName('FileUploadRow')
class FileUploads extends Table {
  TextColumn get uploadId => text()();
  TextColumn get fileName => text()();
  TextColumn get mime => text()();
  IntColumn get sizeBytes => integer().withDefault(const Constant(0))();
  TextColumn get localPath => text().nullable()();
  TextColumn get status => text().withDefault(const Constant('pending'))();
  TextColumn get remoteFileId => text().nullable()();
  TextColumn get error => text().nullable()();
  DateTimeColumn get createdAt => dateTime()();
  /// Raw file bytes so pending uploads survive process restarts.
  BlobColumn get payload => blob().nullable()();

  @override
  Set<Column<Object>> get primaryKey => {uploadId};
}

@DriftDatabase(tables: [CachedEntities, SyncMeta, OutboxOps, FileUploads])
class SfaDatabase extends _$SfaDatabase {
  SfaDatabase(super.e);

  /// Persistent DB for all platforms (Native on IO, Wasm on web via drift_flutter).
  factory SfaDatabase.open() {
    return SfaDatabase(driftDatabase(name: 'sfa_offline_v1'));
  }

  /// In-memory executor for unit tests (VM only).
  factory SfaDatabase.memory(QueryExecutor executor) => SfaDatabase(executor);

  @override
  int get schemaVersion => 3;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onCreate: (m) async => m.createAll(),
        onUpgrade: (m, from, to) async {
          if (from < 2) {
            await m.createTable(fileUploads);
          }
          if (from < 3) {
            await m.addColumn(fileUploads, fileUploads.payload);
          }
        },
      );
}
