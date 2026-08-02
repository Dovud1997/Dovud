import 'package:drift/drift.dart';

/// Web does not use a persistent Drift file DB (no SQLCipher / wasm assets).
/// Entity cache + outbox use encrypted [SecureBlobStore] blobs instead.
/// Calling this is unexpected — prefer blob factories on web.
QueryExecutor openSfaExecutor() {
  throw UnsupportedError(
    'Persistent Drift SQLCipher is not available on web; '
    'use BlobEntityCache / BlobOutboxStore',
  );
}

String sfaConnectionBackendLabel() => 'unavailable (web)';
