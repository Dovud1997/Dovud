/// Abstraction over platform push token providers (FCM / APNs / stub).
abstract class PushTokenSource {
  Future<String?> getToken();

  /// Emits when the platform refreshes the token (FCM). Null when unsupported.
  Stream<String>? get onTokenRefresh;
}

/// Default local stub used when FCM is unavailable (web / missing Firebase config).
class StubPushTokenSource implements PushTokenSource {
  StubPushTokenSource(this._deviceId);

  final Future<String> Function() _deviceId;

  @override
  Future<String?> getToken() async {
    final id = await _deviceId();
    return 'stub-push-$id';
  }

  @override
  Stream<String>? get onTokenRefresh => null;
}
