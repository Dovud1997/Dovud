import 'package:sfa_app/core/offline/entity_cache.dart';
import 'package:sfa_app/core/offline/entity_cache_factory_io.dart'
    if (dart.library.html) 'package:sfa_app/core/offline/entity_cache_factory_web.dart'
    as impl;

/// Opens the platform-appropriate [EntityCache] (SQLite on IO, blob on web).
EntityCache createEntityCache() => impl.createEntityCache();

String entityCacheBackendLabel() => impl.entityCacheBackendLabel();
