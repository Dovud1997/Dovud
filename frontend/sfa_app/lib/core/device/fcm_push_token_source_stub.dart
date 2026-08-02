import 'package:sfa_app/core/device/push_token_source.dart';

/// Web / no-Firebase: nothing to initialize.
Future<bool> ensureFirebaseInitialized() async => false;

PushTokenSource createPushTokenSource(Future<String> Function() deviceId) {
  return StubPushTokenSource(deviceId);
}
