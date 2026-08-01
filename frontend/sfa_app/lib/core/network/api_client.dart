import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/config/app_config.dart';

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient(baseUrl: AppConfig.current.apiBaseUrl);
});

class ApiClient {
  ApiClient({required String baseUrl})
      : dio = Dio(
          BaseOptions(
            baseUrl: baseUrl,
            connectTimeout: const Duration(seconds: 15),
            receiveTimeout: const Duration(seconds: 30),
            headers: {'Content-Type': 'application/json', 'Accept': 'application/json'},
          ),
        );

  final Dio dio;
  String? _accessToken;

  void setAccessToken(String? token) {
    _accessToken = token;
    if (token == null) {
      dio.options.headers.remove('Authorization');
    } else {
      dio.options.headers['Authorization'] = 'Bearer $token';
    }
  }

  String? get accessToken => _accessToken;

  Future<Map<String, dynamic>> get(String path, {Map<String, dynamic>? query}) async {
    final res = await dio.get(path, queryParameters: query);
    return Map<String, dynamic>.from(res.data as Map);
  }

  Future<Map<String, dynamic>> post(String path, {Object? data}) async {
    final res = await dio.post(path, data: data);
    return Map<String, dynamic>.from(res.data as Map);
  }

  Future<Map<String, dynamic>> put(String path, {Object? data}) async {
    final res = await dio.put(path, data: data);
    return Map<String, dynamic>.from(res.data as Map);
  }

  Future<Map<String, dynamic>> patch(String path, {Object? data}) async {
    final res = await dio.patch(path, data: data);
    return Map<String, dynamic>.from(res.data as Map);
  }
}
