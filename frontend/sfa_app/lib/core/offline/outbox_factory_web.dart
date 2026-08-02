import 'package:sfa_app/core/offline/blob_outbox_store.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';

/// Web Admin: AES-encrypted blob outbox.
OutboxStore createOutboxStore() => BlobOutboxStore();

String outboxBackendLabel() => 'SecureBlob outbox';
