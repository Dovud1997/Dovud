import 'package:sfa_app/core/offline/drift/drift_entity_cache.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Mobile / desktop: Drift SQLite (`cached_entities` + `sync_meta`).
EntityCache createEntityCache() => DriftEntityCache(sharedSfaDatabase());

String entityCacheBackendLabel() => 'Drift SQLite';
