import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final identityRepositoryProvider = Provider<IdentityRepository>((ref) {
  return IdentityRepository(ref.watch(apiClientProvider));
});

class IdentityRepository {
  IdentityRepository(this._api);
  final ApiClient _api;

  Future<List<Map<String, dynamic>>> listUsers({int page = 1, int perPage = 50}) async {
    final envelope = await _api.get('/users', query: {'page': page, 'per_page': perPage});
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> createUser(Map<String, dynamic> body) async {
    final envelope = await _api.post('/users', data: body);
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<Map<String, dynamic>> updateUser(String id, Map<String, dynamic> body) async {
    final envelope = await _api.patch('/users/$id', data: body);
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<void> assignRoles(String userId, List<String> roleIds) async {
    await _api.put('/users/$userId/roles', data: {'role_ids': roleIds});
  }

  Future<List<Map<String, dynamic>>> listRoles() async {
    final envelope = await _api.get('/roles');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }

  Future<Map<String, dynamic>> createRole(Map<String, dynamic> body) async {
    final envelope = await _api.post('/roles', data: body);
    return Map<String, dynamic>.from(envelope['data'] as Map? ?? const {});
  }

  Future<void> setRolePermissions(String roleId, List<String> codes) async {
    await _api.put('/roles/$roleId/permissions', data: {'permission_codes': codes});
  }

  Future<List<Map<String, dynamic>>> listPermissions() async {
    final envelope = await _api.get('/permissions');
    final data = envelope['data'] as List<dynamic>? ?? const [];
    return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
  }
}
