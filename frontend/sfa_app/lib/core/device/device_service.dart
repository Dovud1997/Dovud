import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:sfa_app/core/device/push_token_source.dart';
import 'package:sfa_app/core/network/api_client.dart';

final deviceServiceProvider = Provider<DeviceService>((ref) {
  return DeviceService(ref.watch(apiClientProvider));
});

class DeviceService {
  DeviceService(
    this._api, {
    FlutterSecureStorage? storage,
    PushTokenSource? pushTokenSource,
  })  : _storage = storage ?? const FlutterSecureStorage(),
        _pushTokenSource = pushTokenSource;

  final ApiClient _api;
  final FlutterSecureStorage _storage;
  PushTokenSource? _pushTokenSource;
  static const _deviceKey = 'sfa_device_id_v1';

  void setPushTokenSource(PushTokenSource source) {
    _pushTokenSource = source;
  }

  Future<String> deviceId() async {
    final existing = await _storage.read(key: _deviceKey);
    if (existing != null && existing.isNotEmpty) return existing;
    final rand = Random.secure();
    final id = List.generate(16, (_) => rand.nextInt(256))
        .map((b) => b.toRadixString(16).padLeft(2, '0'))
        .join();
    await _storage.write(key: _deviceKey, value: id);
    return id;
  }

  String get platform {
    if (kIsWeb) return 'web';
    return defaultTargetPlatform.name;
  }

  PushTokenSource get _source =>
      _pushTokenSource ?? StubPushTokenSource(deviceId);

  /// Resolves FCM/APNs token when a real [PushTokenSource] is injected;
  /// otherwise returns a deterministic stub token.
  Future<String> pushToken() async {
    final token = await _source.getToken();
    if (token != null && token.trim().isNotEmpty) return token.trim();
    final id = await deviceId();
    return 'stub-push-$id';
  }

  @Deprecated('Use pushToken()')
  Future<String> pushTokenStub() => pushToken();

  Future<void> register({String? appVersion}) async {
    final id = await deviceId();
    final token = await pushToken();
    await _api.post('/auth/devices', data: {
      'device_id': id,
      'platform': platform,
      'push_token': token,
      if (appVersion != null) 'app_version': appVersion,
    });
  }
}
