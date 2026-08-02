import Flutter
import UIKit
import workmanager

@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)

    // Allow plugins (secure storage, path_provider, …) inside Workmanager isolates.
    WorkmanagerPlugin.setPluginRegistrantCallback { registry in
      GeneratedPluginRegistrant.register(with: registry)
    }

    // BGTaskScheduler identifier — must match Info.plist + Dart constant.
    WorkmanagerPlugin.registerTask(withIdentifier: "com.example.sfaApp.backgroundSync")

    UIApplication.shared.setMinimumBackgroundFetchInterval(TimeInterval(60 * 15))

    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
