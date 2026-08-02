import 'package:workmanager/workmanager.dart';
import 'package:sfa_app/core/offline/background_sync_runner.dart';

export 'background_sync_runner.dart' show readOsBackgroundSyncSubtitle;

/// Android WorkManager unique name + task type for periodic sync.
const kBackgroundSyncUniqueName = 'sfa.background.sync';
const kBackgroundSyncTaskName = 'sfaBackgroundSync';

/// iOS BGTaskScheduler identifier (must match Info.plist + AppDelegate).
const kIosBackgroundSyncTaskId = 'com.example.sfaApp.backgroundSync';

bool get isOsBackgroundSyncSupported => true;

String get osBackgroundSyncStatusLabel => 'enabled (~15m OS task)';

@pragma('vm:entry-point')
void backgroundSyncCallbackDispatcher() {
  Workmanager().executeTask((task, inputData) async {
    try {
      // Android periodic uses [kBackgroundSyncTaskName];
      // iOS background fetch uses Workmanager.iOSBackgroundTask.
      return await runBackgroundSyncCycle();
    } catch (_) {
      return false;
    }
  });
}

Future<void> initializeBackgroundSync() async {
  await Workmanager().initialize(
    backgroundSyncCallbackDispatcher,
    isInDebugMode: false,
  );
}

Future<void> scheduleBackgroundSync() async {
  await Workmanager().registerPeriodicTask(
    kBackgroundSyncUniqueName,
    kBackgroundSyncTaskName,
    frequency: const Duration(minutes: 15),
    existingWorkPolicy: ExistingWorkPolicy.keep,
    constraints: Constraints(networkType: NetworkType.connected),
    backoffPolicy: BackoffPolicy.exponential,
    backoffPolicyDelay: const Duration(minutes: 5),
  );
}

Future<void> cancelBackgroundSync() async {
  await Workmanager().cancelByUniqueName(kBackgroundSyncUniqueName);
}
