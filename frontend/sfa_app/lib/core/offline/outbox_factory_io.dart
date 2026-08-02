import 'package:sfa_app/core/offline/drift/drift_outbox_store.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/outbox_store.dart';

OutboxStore createOutboxStore() => DriftOutboxStore(sharedSfaDatabase());

String outboxBackendLabel() => 'Drift outbox_ops';
