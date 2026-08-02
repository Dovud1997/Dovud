import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:sfa_app/core/device/push_token_source.dart';
import 'package:sfa_app/firebase_options.dart';

/// Ensures Firebase is initialized once.
///
/// Order:
/// 1. Already initialized
/// 2. Dart [DefaultFirebaseOptions] when `SFA_FIREBASE_CONFIGURED=true`
/// 3. Native `google-services.json` / `GoogleService-Info.plist`
/// 4. Fail → stub push tokens
Future<bool> ensureFirebaseInitialized() async {
  if (kIsWeb) return false;
  try {
    if (Firebase.apps.isNotEmpty) return true;
    if (DefaultFirebaseOptions.configured) {
      await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
      return true;
    }
    await Firebase.initializeApp();
    return true;
  } catch (_) {
    return false;
  }
}

/// FCM / APNs token source for Android & iOS (APNs via FCM).
class FcmPushTokenSource implements PushTokenSource {
  FcmPushTokenSource(Future<String> Function() deviceId) : _deviceId = deviceId;

  // ignore: unused_field
  final Future<String> Function() _deviceId;

  @override
  Future<String?> getToken() async {
    final ready = await ensureFirebaseInitialized();
    if (!ready) return null;
    try {
      final messaging = FirebaseMessaging.instance;
      await messaging.requestPermission(alert: true, badge: true, sound: true);
      final token = await messaging.getToken();
      if (token != null && token.trim().isNotEmpty) return token.trim();
    } catch (_) {}
    return null;
  }

  @override
  Stream<String>? get onTokenRefresh {
    return Stream.multi((controller) async {
      final ready = await ensureFirebaseInitialized();
      if (!ready) {
        await controller.close();
        return;
      }
      final sub = FirebaseMessaging.instance.onTokenRefresh.listen(
        controller.add,
        onError: controller.addError,
        onDone: controller.close,
      );
      controller.onCancel = sub.cancel;
    });
  }
}

PushTokenSource createPushTokenSource(Future<String> Function() deviceId) {
  if (kIsWeb) return StubPushTokenSource(deviceId);
  return FcmPushTokenSource(deviceId);
}
