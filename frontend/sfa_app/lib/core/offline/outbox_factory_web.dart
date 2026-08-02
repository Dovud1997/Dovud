import 'package:sfa_app/core/offline/blob_outbox_store.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';

OutboxStore createOutboxStore() => BlobOutboxStore();

String outboxBackendLabel() => 'encrypted blob (web)';
