/// Typed local entity cache.
///
/// Implementations: SQLite tables (mobile/desktop) and encrypted blob (web).
abstract class EntityCache {
  Future<List<Map<String, dynamic>>> listEntities(String type);

  Future<void> upsertEntity(String type, Map<String, dynamic> entity);

  Future<void> deleteEntity(String type, String id);

  Future<String?> cursor();

  Future<void> setCursor(String value);
}
