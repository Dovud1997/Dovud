import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/sfa_connection.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Native: Drift EntityCache over SQLCipher-encrypted SQLite.
EntityCache createEntityCache() => DriftEntityCache(sharedSfaDatabase());

String entityCacheBackendLabel() => sfaConnectionBackendLabel();
