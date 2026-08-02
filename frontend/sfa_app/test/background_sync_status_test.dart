import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:sfa_app/core/offline/background_sync_runner.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  test('readOsBackgroundSyncSubtitle waiting when empty', () async {
    SharedPreferences.setMockInitialValues({});
    expect(await readOsBackgroundSyncSubtitle(), 'waiting for first OS run');
  });

  test('readOsBackgroundSyncSubtitle prefers error over ok', () async {
    SharedPreferences.setMockInitialValues({
      kOsBgSyncLastOkKey: '2026-01-01T00:00:00.000Z',
      kOsBgSyncLastErrorKey: 'boom',
    });
    expect(await readOsBackgroundSyncSubtitle(), 'error: boom');
  });

  test('readOsBackgroundSyncSubtitle shows last ok', () async {
    SharedPreferences.setMockInitialValues({
      kOsBgSyncLastOkKey: '2026-01-01T00:00:00.000Z',
    });
    expect(await readOsBackgroundSyncSubtitle(), 'ok · 2026-01-01T00:00:00.000Z');
  });
}
