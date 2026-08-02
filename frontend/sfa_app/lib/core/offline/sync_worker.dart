import 'dart:async';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/background_sync.dart' as bg_sync;
import 'package:sfa_app/core/offline/file_upload_queue.dart';
import 'package:sfa_app/core/offline/gps_queue.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

final syncWorkerProvider = Provider<SyncWorker>((ref) {
  final worker = SyncWorker(ref);
  ref.onDispose(worker.dispose);
  return worker;
});

/// Foreground background-sync: periodic + connectivity + app resume.
///
/// Also schedules an OS WorkManager / BGTask (~15m) while authenticated.
class SyncWorker with WidgetsBindingObserver {
  SyncWorker(this._ref) {
    WidgetsBinding.instance.addObserver(this);
    _timer = Timer.periodic(const Duration(seconds: 45), (_) {
      unawaited(tick(reason: 'timer'));
    });
    _connectivitySub = Connectivity().onConnectivityChanged.listen((results) {
      final online = results.any((r) => r != ConnectivityResult.none);
      if (online) unawaited(tick(reason: 'connectivity'));
    });
    unawaited(_safeScheduleOsBackgroundSync());
  }

  final Ref _ref;
  Timer? _timer;
  StreamSubscription<List<ConnectivityResult>>? _connectivitySub;
  bool _running = false;
  String? lastError;
  DateTime? lastSuccessAt;
  Map<String, dynamic>? lastResult;

  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _timer?.cancel();
    _connectivitySub?.cancel();
    unawaited(_safeCancelOsBackgroundSync());
  }

  Future<void> _safeScheduleOsBackgroundSync() async {
    try {
      await bg_sync.scheduleBackgroundSync();
    } catch (_) {}
  }

  Future<void> _safeCancelOsBackgroundSync() async {
    try {
      await bg_sync.cancelBackgroundSync();
    } catch (_) {}
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      unawaited(tick(reason: 'resume'));
    }
  }

  Future<Map<String, dynamic>?> tick({String reason = 'manual'}) async {
    final session = _ref.read(sessionControllerProvider);
    if (!session.isAuthenticated || session.loading) return null;
    if (_running) return lastResult;
    _running = true;
    try {
      final store = _ref.read(offlineStoreProvider);
      final result = await store.syncCycle();
      Map<String, dynamic> uploads = const {};
      Map<String, dynamic> gps = const {};
      try {
        uploads = await _ref.read(fileUploadQueueProvider).flush();
      } catch (_) {}
      try {
        gps = await _ref.read(gpsQueueProvider).flush();
      } catch (_) {}
      lastResult = {...result, 'uploads': uploads, 'gps': gps, 'reason': reason};
      lastSuccessAt = DateTime.now().toUtc();
      lastError = null;
      return lastResult;
    } catch (e) {
      lastError = e.toString();
      return null;
    } finally {
      _running = false;
    }
  }
}
