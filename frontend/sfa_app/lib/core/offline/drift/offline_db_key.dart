import 'dart:math';
import 'dart:typed_data';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// 256-bit database key stored in [FlutterSecureStorage] for SQLCipher.
class OfflineDbKey {
  OfflineDbKey({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;
  static const storageKey = 'sfa_drift_sqlcipher_key_v1';

  /// Returns a 64-char hex string suitable for `PRAGMA key = "x'…'"`.
  Future<String> hexKey() async {
    final existing = await _storage.read(key: storageKey);
    if (existing != null && existing.length == 64 && isValidSqlCipherHexKey(existing)) {
      return existing;
    }

    final raw = Uint8List(32);
    final rand = Random.secure();
    for (var i = 0; i < raw.length; i++) {
      raw[i] = rand.nextInt(256);
    }
    final hex = raw.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
    await _storage.write(key: storageKey, value: hex);
    return hex;
  }
}

/// Escape-safe PRAGMA key assignment using a raw hex key.
String sqlCipherPragmaKey(String hexKey) => "PRAGMA key = \"x'$hexKey'\";";

/// Validates [hex] looks like 32 random bytes.
bool isValidSqlCipherHexKey(String hex) =>
    hex.length == 64 && RegExp(r'^[0-9a-fA-F]{64}$').hasMatch(hex);
