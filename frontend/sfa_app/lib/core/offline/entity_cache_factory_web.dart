import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Web Admin: Drift Wasm SQLite (same schema as mobile).
EntityCache createEntityCache() => DriftEntityCache(sharedSfaDatabase());

String entityCacheBackendLabel() => 'Drift Wasm';
