/// Typed local entity cache. Current impl is encrypted blob prefs;
/// swap to Drift/sqflite later without changing Sync page call sites.
abstract class EntityCache {
  Future<List<Map<String, dynamic>>> listEntities(String type);

  Future<void> upsertEntity(String type, Map<String, dynamic> entity);

  Future<void> deleteEntity(String type, String id);

  Future<String?> cursor();

  Future<void> setCursor(String value);
}
