// File generated manually as a Firebase options template for white-label builds.
// Replace placeholder values with output from `flutterfire configure`, or keep
// `SFA_FIREBASE_ENABLED=false` and the app will fall back to StubPushTokenSource.
//
// ignore_for_file: lines_longer_than_80_chars

import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

/// Default [FirebaseOptions] for the current platform.
///
/// Set [configured] to true after filling real project values (or regenerate
/// this file with FlutterFire CLI).
class DefaultFirebaseOptions {
  /// Flip to true once real Firebase project credentials are filled in.
  static const bool configured = bool.fromEnvironment(
    'SFA_FIREBASE_CONFIGURED',
    defaultValue: false,
  );

  static FirebaseOptions get currentPlatform {
    if (kIsWeb) {
      return web;
    }
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return android;
      case TargetPlatform.iOS:
        return ios;
      case TargetPlatform.macOS:
        return macos;
      default:
        throw UnsupportedError(
          'DefaultFirebaseOptions are not supported for this platform.',
        );
    }
  }

  static const FirebaseOptions web = FirebaseOptions(
    apiKey: 'REPLACE_WEB_API_KEY',
    appId: '1:000000000000:web:replace_me',
    messagingSenderId: '000000000000',
    projectId: 'sfa-replace-me',
    authDomain: 'sfa-replace-me.firebaseapp.com',
    storageBucket: 'sfa-replace-me.appspot.com',
  );

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'REPLACE_ANDROID_API_KEY',
    appId: '1:000000000000:android:replace_me',
    messagingSenderId: '000000000000',
    projectId: 'sfa-replace-me',
    storageBucket: 'sfa-replace-me.appspot.com',
  );

  static const FirebaseOptions ios = FirebaseOptions(
    apiKey: 'REPLACE_IOS_API_KEY',
    appId: '1:000000000000:ios:replace_me',
    messagingSenderId: '000000000000',
    projectId: 'sfa-replace-me',
    storageBucket: 'sfa-replace-me.appspot.com',
    iosBundleId: 'com.example.sfaApp',
  );

  static const FirebaseOptions macos = FirebaseOptions(
    apiKey: 'REPLACE_MACOS_API_KEY',
    appId: '1:000000000000:ios:replace_me',
    messagingSenderId: '000000000000',
    projectId: 'sfa-replace-me',
    storageBucket: 'sfa-replace-me.appspot.com',
    iosBundleId: 'com.example.sfaApp',
  );
}
