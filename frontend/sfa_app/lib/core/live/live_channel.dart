import 'dart:async';
import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/config/app_config.dart';
import 'package:sfa_app/core/offline/sync_worker.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

final liveChannelProvider = Provider<LiveChannel>((ref) {
  final channel = LiveChannel(ref);
  ref.onDispose(channel.dispose);
  ref.listen(sessionControllerProvider, (prev, next) {
    if (next.isAuthenticated) {
      unawaited(channel.connect());
    } else {
      channel.disconnect();
    }
  });
  return channel;
});

/// Client for `/ws/v1` live events (sync.invalidate, domain updates, ping/pong).
class LiveChannel {
  LiveChannel(this._ref);

  final Ref _ref;
  WebSocketChannel? _channel;
  StreamSubscription? _sub;
  StreamSubscription? _eventsSub;
  final _controller = StreamController<Map<String, dynamic>>.broadcast();
  String? lastError;
  bool connected = false;
  DateTime? lastInvalidateAt;

  Stream<Map<String, dynamic>> get events => _controller.stream;

  Future<void> connect() async {
    await disconnect();
    final session = _ref.read(sessionControllerProvider).session;
    if (session == null) return;
    try {
      final base = AppConfig.current.apiBaseUrl.replaceFirst(RegExp(r'/api/v1/?$'), '');
      final wsBase = base
          .replaceFirst('https://', 'wss://')
          .replaceFirst('http://', 'ws://');
      final uri = Uri.parse('$wsBase/ws/v1').replace(queryParameters: {
        'token': session.accessToken,
      });
      final ch = WebSocketChannel.connect(uri);
      _channel = ch;
      connected = true;
      lastError = null;
      _sub = ch.stream.listen(
        (raw) {
          try {
            final map = jsonDecode('$raw') as Map<String, dynamic>;
            _controller.add(map);
          } catch (_) {}
        },
        onError: (e) {
          lastError = e.toString();
          connected = false;
        },
        onDone: () {
          connected = false;
        },
      );
      _eventsSub = _controller.stream.listen(_onEvent);
      ch.sink.add(jsonEncode({'type': 'ping'}));
    } catch (e) {
      lastError = e.toString();
      connected = false;
      if (kDebugMode) {
        // ignore: avoid_print
        print('ws connect failed: $e');
      }
    }
  }

  void _onEvent(Map<String, dynamic> map) {
    final type = map['type']?.toString() ?? '';
    if (type == 'sync.invalidate' ||
        type == 'order.updated' ||
        type == 'visit.updated' ||
        type == 'notification.created' ||
        type == 'product.updated' ||
        type == 'return.updated') {
      lastInvalidateAt = DateTime.now().toUtc();
      unawaited(_ref.read(syncWorkerProvider).tick(reason: 'ws:$type'));
    }
  }

  Future<void> disconnect() async {
    await _eventsSub?.cancel();
    _eventsSub = null;
    await _sub?.cancel();
    _sub = null;
    await _channel?.sink.close();
    _channel = null;
    connected = false;
  }

  void dispose() {
    unawaited(disconnect());
    unawaited(_controller.close());
  }
}
