import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/app.dart';
import 'package:sfa_app/core/device/fcm_push_token_source_io.dart'
    if (dart.library.html) 'package:sfa_app/core/device/fcm_push_token_source_stub.dart'
    as push_boot;

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  if (!kIsWeb) {
    // Best-effort Firebase bootstrap; missing google-services / GoogleService-Info
    // falls back to StubPushTokenSource inside DeviceService.
    await push_boot.ensureFirebaseInitialized();
  }
  runApp(const ProviderScope(child: SfaApp()));
}
