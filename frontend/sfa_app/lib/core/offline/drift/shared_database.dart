import 'package:sfa_app/core/offline/drift/sfa_database.dart';

SfaDatabase? _shared;

/// Shared Drift DB for entity cache + outbox + queues on IO (SQLCipher).
SfaDatabase sharedSfaDatabase() => _shared ??= SfaDatabase.open();

/// Test helper — replaces the shared instance.
void debugSetSharedSfaDatabase(SfaDatabase? db) {
  _shared = db;
}
