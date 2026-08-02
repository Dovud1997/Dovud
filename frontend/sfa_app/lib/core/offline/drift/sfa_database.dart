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

@DriftDatabase(tables: [CachedEntities, SyncMeta, OutboxOps])
class SfaDatabase extends _$SfaDatabase {
  SfaDatabase(super.e);

  /// Persistent DB for all platforms (Native on IO, Wasm on web via drift_flutter).
  factory SfaDatabase.open() {
    return SfaDatabase(driftDatabase(name: 'sfa_offline_v1'));
  }

  /// In-memory executor for unit tests (VM only).
  factory SfaDatabase.memory(QueryExecutor executor) => SfaDatabase(executor);

  @override
  int get schemaVersion => 1;
}
