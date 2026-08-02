import 'package:drift/drift.dart';
import 'package:sfa_app/core/offline/drift/sfa_connection.dart';

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

@DataClassName('GpsPendingRow')
class GpsPending extends Table {
  TextColumn get pointId => text()();
  TextColumn get agentId => text()();
  TextColumn get visitId => text().nullable()();
  RealColumn get lat => real()();
  RealColumn get lng => real()();
  RealColumn get accuracy => real().nullable()();
  DateTimeColumn get recordedAt => dateTime()();
  TextColumn get status => text().withDefault(const Constant('pending'))();
  TextColumn get error => text().nullable()();
  DateTimeColumn get createdAt => dateTime()();

  @override
  Set<Column<Object>> get primaryKey => {pointId};
}

@DriftDatabase(tables: [CachedEntities, SyncMeta, OutboxOps, FileUploads, GpsPending])
class SfaDatabase extends _$SfaDatabase {
  SfaDatabase(super.e);

  /// Persistent encrypted DB on native (SQLCipher). Web uses blob stores instead.
  factory SfaDatabase.open() {
    return SfaDatabase(openSfaExecutor());
  }

  /// In-memory executor for unit tests (VM only).
  factory SfaDatabase.memory(QueryExecutor executor) => SfaDatabase(executor);

  @override
  int get schemaVersion => 4;

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
          if (from < 4) {
            await m.createTable(gpsPending);
          }
        },
      );
}
