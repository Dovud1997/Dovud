import 'package:sfa_app/core/device/fcm_push_token_source_io.dart'
    if (dart.library.html) 'package:sfa_app/core/device/fcm_push_token_source_stub.dart'
    as impl;
import 'package:sfa_app/core/device/push_token_source.dart';

/// Platform push token source (FCM/APNs on mobile, stub on web).
PushTokenSource createPushTokenSource(Future<String> Function() deviceId) =>
    impl.createPushTokenSource(deviceId);
