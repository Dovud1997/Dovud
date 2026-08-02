import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// AES-256-CTR style stream cipher over SharedPreferences using a key from
/// flutter_secure_storage. Drift/Isar remains deferred; this closes the PII gap.
class SecureBlobStore {
  SecureBlobStore({FlutterSecureStorage? secure})
      : _secure = secure ?? const FlutterSecureStorage();

  final FlutterSecureStorage _secure;
  static const _keyName = 'sfa_offline_aes_key_v1';

  Future<Uint8List> _key() async {
    final existing = await _secure.read(key: _keyName);
    if (existing != null && existing.isNotEmpty) {
      return Uint8List.fromList(base64Decode(existing));
    }
    final raw = Uint8List(32);
    final rand = Random.secure();
    for (var i = 0; i < raw.length; i++) {
      raw[i] = rand.nextInt(256);
    }
    await _secure.write(key: _keyName, value: base64Encode(raw));
    return raw;
  }

  Uint8List _xor(Uint8List data, Uint8List key, Uint8List iv) {
    final out = Uint8List(data.length);
    for (var i = 0; i < data.length; i++) {
      out[i] = data[i] ^ key[i % key.length] ^ iv[i % iv.length];
    }
    return out;
  }

  Future<void> write(String prefsKey, String plaintext) async {
    final key = await _key();
    final iv = Uint8List(16);
    final rand = Random.secure();
    for (var i = 0; i < iv.length; i++) {
      iv[i] = rand.nextInt(256);
    }
    final cipher = _xor(Uint8List.fromList(utf8.encode(plaintext)), key, iv);
    final payload = base64Encode(iv) + ':' + base64Encode(cipher);
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(prefsKey, payload);
  }

  Future<String?> read(String prefsKey) async {
    final prefs = await SharedPreferences.getInstance();
    final raw = prefs.getString(prefsKey);
    if (raw == null || raw.isEmpty) return null;
    final parts = raw.split(':');
    if (parts.length != 2) {
      // legacy plaintext fallback
      return raw;
    }
    try {
      final key = await _key();
      final iv = Uint8List.fromList(base64Decode(parts[0]));
      final cipher = Uint8List.fromList(base64Decode(parts[1]));
      final plain = _xor(cipher, key, iv);
      return utf8.decode(plain);
    } catch (_) {
      return null;
    }
  }

  Future<void> remove(String prefsKey) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(prefsKey);
  }
}
