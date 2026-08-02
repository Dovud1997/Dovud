import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:sfa_app/core/config/app_config.dart';
import 'package:sfa_app/core/device/device_service.dart';
import 'package:sfa_app/core/network/api_client.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/core/offline/file_upload_queue.dart';
import 'package:sfa_app/core/offline/gps_queue.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/features/auth/data/auth_repository.dart';
import 'package:sfa_app/features/documents/data/documents_repository.dart';
import 'package:sfa_app/features/fieldforce/data/fieldforce_repository.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

/// SharedPreferences keys written by the OS background sync isolate.
const kOsBgSyncLastOkKey = 'os_bg_sync_last_ok';
const kOsBgSyncLastErrorKey = 'os_bg_sync_last_error';
const kOsBgSyncLastSummaryKey = 'os_bg_sync_last_summary';

/// Headless sync cycle for Workmanager / BGTask isolates (no Riverpod).
///
/// Returns `true` when work completed or was skipped (no session);
/// `false` when a retryable failure occurred.
Future<bool> runBackgroundSyncCycle() async {
  const storage = FlutterSecureStorage();
  final access = await storage.read(key: AuthRepository.accessTokenKey);
  final refresh = await storage.read(key: AuthRepository.refreshTokenKey);
  if ((access == null || access.isEmpty) && (refresh == null || refresh.isEmpty)) {
    await _recordSkip('no session');
    return true;
  }

  final api = ApiClient(baseUrl: AppConfig.current.apiBaseUrl);
  final devices = DeviceService(api, storage: storage);
  final deviceId = await devices.deviceId();

  try {
    final token = await _ensureAccessToken(
      api: api,
      storage: storage,
      access: access,
      refresh: refresh,
      deviceId: deviceId,
    );
    if (token == null) {
      await _recordError('auth expired');
      return true; // do not retry forever when logged out / refresh dead
    }
    api.setAccessToken(token);

    final sync = SyncRepository(api, devices);
    final store = OfflineStore(sync, LocalOutbox(sync));
    final docs = DocumentsRepository(api);
    final uploads = FileUploadQueue(
      sharedSfaDatabase(),
      uploadBytes: docs.uploadBytes,
    );
    final gps = GpsQueue(sharedSfaDatabase(), FieldForceRepository(api));

    final cycle = await store.syncCycle(deviceId: deviceId);
    Map<String, dynamic> uploadRes = const {};
    Map<String, dynamic> gpsRes = const {};
    try {
      uploadRes = await uploads.flush();
    } catch (_) {}
    try {
      gpsRes = await gps.flush();
    } catch (_) {}

    final summary =
        'flush=${cycle['flush']} pull=${cycle['pull']} uploads=$uploadRes gps=$gpsRes';
    await _recordOk(summary);
    return true;
  } catch (e) {
    await _recordError(e.toString());
    return false;
  }
}

Future<String?> _ensureAccessToken({
  required ApiClient api,
  required FlutterSecureStorage storage,
  required String? access,
  required String? refresh,
  required String deviceId,
}) async {
  if (refresh != null && refresh.isNotEmpty) {
    try {
      final envelope = await api.post('/auth/refresh', data: {
        'refresh_token': refresh,
        'device_id': deviceId,
      });
      final data = Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
      final nextAccess = data['access_token']?.toString();
      final nextRefresh = data['refresh_token']?.toString();
      if (nextAccess != null && nextAccess.isNotEmpty) {
        await storage.write(key: AuthRepository.accessTokenKey, value: nextAccess);
        if (nextRefresh != null && nextRefresh.isNotEmpty) {
          await storage.write(key: AuthRepository.refreshTokenKey, value: nextRefresh);
        }
        return nextAccess;
      }
    } catch (_) {
      // fall through to existing access token
    }
  }
  if (access != null && access.isNotEmpty) return access;
  return null;
}

Future<void> _recordOk(String summary) async {
  final prefs = await SharedPreferences.getInstance();
  await prefs.setString(kOsBgSyncLastOkKey, DateTime.now().toUtc().toIso8601String());
  await prefs.setString(kOsBgSyncLastSummaryKey, summary);
  await prefs.remove(kOsBgSyncLastErrorKey);
}

Future<void> _recordError(String error) async {
  final prefs = await SharedPreferences.getInstance();
  await prefs.setString(kOsBgSyncLastErrorKey, error);
}

Future<void> _recordSkip(String reason) async {
  final prefs = await SharedPreferences.getInstance();
  await prefs.setString(kOsBgSyncLastSummaryKey, 'skipped: $reason');
}

/// Reads last OS background sync status for the Sync center UI.
Future<String> readOsBackgroundSyncSubtitle() async {
  final prefs = await SharedPreferences.getInstance();
  final err = prefs.getString(kOsBgSyncLastErrorKey);
  if (err != null && err.isNotEmpty) return 'error: $err';
  final ok = prefs.getString(kOsBgSyncLastOkKey);
  if (ok != null && ok.isNotEmpty) return 'ok · $ok';
  return 'waiting for first OS run';
}
