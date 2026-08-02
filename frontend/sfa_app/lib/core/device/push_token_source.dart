/// Abstraction over platform push token providers (FCM / APNs / stub).
abstract class PushTokenSource {
  Future<String?> getToken();
}

/// Default local stub used until firebase_messaging is wired.
class StubPushTokenSource implements PushTokenSource {
  StubPushTokenSource(this._deviceId);

  final Future<String> Function() _deviceId;

  @override
  Future<String?> getToken() async {
    final id = await _deviceId();
    return 'stub-push-$id';
  }
}
