// Web / unsupported platforms — OS background sync is a no-op.

bool get isOsBackgroundSyncSupported => false;

String get osBackgroundSyncStatusLabel => 'unavailable (web)';

Future<void> initializeBackgroundSync() async {}

Future<void> scheduleBackgroundSync() async {}

Future<void> cancelBackgroundSync() async {}

Future<String> readOsBackgroundSyncSubtitle() async => 'unavailable (web)';
