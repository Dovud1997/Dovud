import 'package:sfa_app/core/offline/blob_entity_cache.dart';
import 'package:sfa_app/core/offline/entity_cache.dart';

/// Web Admin: AES-encrypted SharedPreferences blob (no SQLCipher / Drift file).
EntityCache createEntityCache() => BlobEntityCache();

String entityCacheBackendLabel() => 'SecureBlob AES';
