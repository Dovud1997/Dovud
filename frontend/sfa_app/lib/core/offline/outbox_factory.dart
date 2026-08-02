import 'package:sfa_app/core/offline/outbox_factory_io.dart'
    if (dart.library.html) 'package:sfa_app/core/offline/outbox_factory_web.dart'
    as impl;
import 'package:sfa_app/core/offline/outbox_store.dart';

OutboxStore createOutboxStore() => impl.createOutboxStore();

String outboxBackendLabel() => impl.outboxBackendLabel();
