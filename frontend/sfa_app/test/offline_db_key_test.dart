import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/core/offline/drift/offline_db_key.dart';

void main() {
  test('sqlCipherPragmaKey formats raw hex key', () {
    const hex = '00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff';
    expect(sqlCipherPragmaKey(hex), "PRAGMA key = \"x'$hex'\";");
  });

  test('isValidSqlCipherHexKey', () {
    expect(
      isValidSqlCipherHexKey(
        '00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff',
      ),
      isTrue,
    );
    expect(isValidSqlCipherHexKey('short'), isFalse);
    expect(isValidSqlCipherHexKey('gg' * 32), isFalse);
  });
}
