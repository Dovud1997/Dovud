import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// All platforms: Drift EntityCache (Wasm on web via drift_flutter).
EntityCache createEntityCache() => DriftEntityCache(sharedSfaDatabase());

String entityCacheBackendLabel() => 'Drift SQLite';
