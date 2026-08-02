import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/device/device_service.dart';
import 'package:sfa_app/core/network/api_client.dart';

final syncRepositoryProvider = Provider<SyncRepository>((ref) {
  return SyncRepository(
    ref.watch(apiClientProvider),
    ref.watch(deviceServiceProvider),
  );
});

class SyncConflict {
  SyncConflict({
    required this.id,
    required this.entityType,
    required this.entityId,
    required this.clientOpId,
    required this.baseVersion,
    required this.serverVersion,
    required this.clientPayload,
    required this.serverPayload,
    required this.status,
    this.resolution,
  });

  final String id;
  final String entityType;
  final String entityId;
  final String clientOpId;
  final int baseVersion;
  final int serverVersion;
  final Map<String, dynamic> clientPayload;
  final Map<String, dynamic> serverPayload;
  final String status;
  final String? resolution;

  factory SyncConflict.fromJson(Map<String, dynamic> json) {
    Map<String, dynamic> asMap(dynamic v) {
      if (v is Map) return Map<String, dynamic>.from(v);
      return const {};
    }

    return SyncConflict(
      id: json['id']?.toString() ?? '',
      entityType: json['entity_type']?.toString() ?? '',
      entityId: json['entity_id']?.toString() ?? '',
      clientOpId: json['client_op_id']?.toString() ?? '',
      baseVersion: (json['base_version'] as num?)?.toInt() ?? 0,
      serverVersion: (json['server_version'] as num?)?.toInt() ?? 0,
      clientPayload: asMap(json['client_payload']),
      serverPayload: asMap(json['server_payload']),
      status: json['status']?.toString() ?? 'open',
      resolution: json['resolution']?.toString(),
    );
  }
}

class SyncRepository {
  SyncRepository(this._api, this._devices);

  final ApiClient _api;
  final DeviceService _devices;

  Future<String> deviceId() => _devices.deviceId();

  Future<Map<String, dynamic>> status({String? deviceId}) async {
    final id = deviceId ?? await this.deviceId();
    final envelope = await _api.get('/sync/status', query: {'device_id': id});
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> bootstrap({String? deviceId}) async {
    final id = deviceId ?? await this.deviceId();
    final envelope = await _api.post('/sync/bootstrap', data: {
      'device_id': id,
      'platform': _devices.platform,
      'app_version': '0.1.0',
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> pull({String? deviceId, String cursor = ''}) async {
    final id = deviceId ?? await this.deviceId();
    final envelope = await _api.get('/sync/pull', query: {
      'device_id': id,
      if (cursor.isNotEmpty) 'cursor': cursor,
      'limit': 100,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> push({
    String? deviceId,
    List<Map<String, dynamic>> ops = const [],
  }) async {
    final id = deviceId ?? await this.deviceId();
    final envelope = await _api.post('/sync/push', data: {
      'device_id': id,
      'ops': ops,
    });
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<List<SyncConflict>> listConflicts({String? deviceId}) async {
    final id = deviceId ?? await this.deviceId();
    final envelope = await _api.get('/sync/conflicts', query: {'device_id': id});
    final data = envelope['data'];
    final list = data is List ? data : const [];
    return list
        .map((e) => SyncConflict.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
  }

  Future<SyncConflict> resolveConflict({
    required String conflictId,
    required String resolution,
  }) async {
    final envelope = await _api.post('/sync/conflicts/$conflictId/resolve', data: {
      'resolution': resolution,
    });
    return SyncConflict.fromJson(Map<String, dynamic>.from(envelope['data'] as Map));
  }
}
