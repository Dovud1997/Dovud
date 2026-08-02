import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:sfa_app/core/device/device_service.dart';
import 'package:sfa_app/core/network/api_client.dart';
import 'package:sfa_app/features/auth/domain/user.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    ref.watch(apiClientProvider),
    ref.watch(deviceServiceProvider),
  );
});

class AuthSession {
  const AuthSession({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });

  final String accessToken;
  final String refreshToken;
  final AuthUser user;
}

class AuthRepository {
  AuthRepository(this._api, this._devices, {FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  final ApiClient _api;
  final DeviceService _devices;
  final FlutterSecureStorage _storage;

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';

  Future<AuthSession> login({
    required String tenantCode,
    required String email,
    required String password,
  }) async {
    final deviceId = await _devices.deviceId();
    final envelope = await _api.post('/auth/login', data: {
      'tenant_code': tenantCode,
      'email': email,
      'password': password,
      'device_id': deviceId,
      'platform': _devices.platform,
    });
    final data = Map<String, dynamic>.from(envelope['data'] as Map);
    final session = AuthSession(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
      user: AuthUser.fromJson(Map<String, dynamic>.from(data['user'] as Map)),
    );
    await _persist(session);
    _api.setAccessToken(session.accessToken);
    await _safeRegisterDevice();
    return session;
  }

  Future<AuthSession?> restore() async {
    final access = await _storage.read(key: _accessKey);
    final refresh = await _storage.read(key: _refreshKey);
    if (access == null || refresh == null) return null;
    _api.setAccessToken(access);
    try {
      final envelope = await _api.get('/auth/me');
      final user = AuthUser.fromJson(Map<String, dynamic>.from(envelope['data'] as Map));
      await _safeRegisterDevice();
      return AuthSession(accessToken: access, refreshToken: refresh, user: user);
    } catch (_) {
      return refreshSession(refresh);
    }
  }

  Future<AuthSession?> refreshSession(String refreshToken) async {
    try {
      final deviceId = await _devices.deviceId();
      final envelope = await _api.post('/auth/refresh', data: {
        'refresh_token': refreshToken,
        'device_id': deviceId,
      });
      final data = Map<String, dynamic>.from(envelope['data'] as Map);
      final session = AuthSession(
        accessToken: data['access_token'] as String,
        refreshToken: data['refresh_token'] as String,
        user: AuthUser.fromJson(Map<String, dynamic>.from(data['user'] as Map)),
      );
      await _persist(session);
      _api.setAccessToken(session.accessToken);
      await _safeRegisterDevice();
      return session;
    } catch (_) {
      await clear();
      return null;
    }
  }

  Future<void> logout(String? refreshToken) async {
    try {
      if (refreshToken != null) {
        await _api.post('/auth/logout', data: {'refresh_token': refreshToken});
      }
    } catch (_) {}
    await clear();
  }

  Future<void> _safeRegisterDevice() async {
    try {
      await _devices.register();
    } catch (_) {}
  }

  Future<void> _persist(AuthSession session) async {
    await _storage.write(key: _accessKey, value: session.accessToken);
    await _storage.write(key: _refreshKey, value: session.refreshToken);
  }

  Future<void> clear() async {
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
    _api.setAccessToken(null);
  }
}
