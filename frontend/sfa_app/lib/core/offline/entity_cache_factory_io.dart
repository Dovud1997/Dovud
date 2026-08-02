import 'package:sfa_app/core/offline/entity_cache.dart';
import 'package:sfa_app/core/offline/sqlite_entity_cache.dart';

/// Mobile / desktop: SQLite entity tables (Drift-shaped schema).
EntityCache createEntityCache() => SqliteEntityCache();

String entityCacheBackendLabel() => 'SQLite tables';
