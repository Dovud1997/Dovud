import 'dart:io';

import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:sqlcipher_flutter_libs/sqlcipher_flutter_libs.dart';
import 'package:sqlite3/open.dart';
import 'package:sqlite3/sqlite3.dart';
import 'package:sfa_app/core/offline/drift/offline_db_key.dart';

const kEncryptedDbFileName = 'sfa_offline_v2_enc.sqlite';
const kLegacyPlainDbFileName = 'sfa_offline_v1.sqlite';

bool _cipherConfigured = false;

Future<void> _ensureSqlCipherConfigured() async {
  if (_cipherConfigured) return;
  if (Platform.isAndroid) {
    open.overrideFor(OperatingSystem.android, openCipherOnAndroid);
    await applyWorkaroundToOpenSqlCipherOnOldAndroidVersions();
  }
  // Prefer app-private temp dir (Android sandbox).
  try {
    sqlite3.tempDirectory = (await getTemporaryDirectory()).path;
  } catch (_) {}
  _cipherConfigured = true;
}

/// Opens an encrypted Drift executor (SQLCipher) for native platforms.
QueryExecutor openSfaExecutor() {
  return DatabaseConnection.delayed(Future(() async {
    await _ensureSqlCipherConfigured();
    final hex = await OfflineDbKey().hexKey();
    final dir = await getApplicationDocumentsDirectory();
    final file = File(p.join(dir.path, kEncryptedDbFileName));

    // Drop legacy plaintext DB so we never mix schemas/keys.
    try {
      final legacy = File(p.join(dir.path, kLegacyPlainDbFileName));
      if (await legacy.exists()) {
        await legacy.delete();
      }
    } catch (_) {}

    return NativeDatabase.createBackgroundConnection(
      file,
      isolateSetup: () async {
        // Background isolate must also point sqlite3 at SQLCipher on Android.
        if (Platform.isAndroid) {
          open.overrideFor(OperatingSystem.android, openCipherOnAndroid);
        }
      },
      setup: (rawDb) {
        rawDb.execute(sqlCipherPragmaKey(hex));
        // Fail loud if the linked library is plain sqlite3 (PRAGMA key is a no-op).
        final rows = rawDb.select('PRAGMA cipher_version;');
        if (rows.isEmpty) {
          throw StateError(
            'SQLCipher is not available — check sqlcipher_flutter_libs linkage',
          );
        }
      },
    );
  }));
}

String sfaConnectionBackendLabel() => 'Drift SQLCipher';
