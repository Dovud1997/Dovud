import 'package:sfa_app/core/offline/blob_entity_cache.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Web Admin: encrypted blob prefs (no sqflite on browser builds).
EntityCache createEntityCache() => BlobEntityCache();

String entityCacheBackendLabel() => 'encrypted blob (web)';
