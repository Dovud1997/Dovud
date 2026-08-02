import 'dart:convert';

/// Builds a merged `lines` array from server/client payloads using per-line picks.
///
/// [pickClient] maps line keys → true if the client line should win.
/// Keys are typically `product_id` (see [lineKey]).
List<Map<String, dynamic>> mergeOrderLines({
  required dynamic serverLines,
  required dynamic clientLines,
  required Map<String, bool> pickClient,
}) {
  final server = asLineMaps(serverLines);
  final client = asLineMaps(clientLines);
  final serverByKey = {for (final l in server) lineKey(l): l};
  final clientByKey = {for (final l in client) lineKey(l): l};
  final keys = {...serverByKey.keys, ...clientByKey.keys}.toList()..sort();

  final out = <Map<String, dynamic>>[];
  for (final key in keys) {
    final useClient = pickClient[key] ?? clientByKey.containsKey(key);
    if (useClient) {
      final line = clientByKey[key];
      if (line != null) out.add(Map<String, dynamic>.from(line));
    } else {
      final line = serverByKey[key];
      if (line != null) out.add(Map<String, dynamic>.from(line));
    }
  }
  return out;
}

/// Stable key for a line: product_id, else id, else json fingerprint.
String lineKey(Map<String, dynamic> line) {
  final pid = line['product_id']?.toString();
  if (pid != null && pid.isNotEmpty) return 'p:$pid';
  final id = line['id']?.toString();
  if (id != null && id.isNotEmpty) return 'i:$id';
  return 'j:${jsonEncode(line)}';
}

List<Map<String, dynamic>> asLineMaps(dynamic raw) {
  if (raw is! List) return const [];
  final out = <Map<String, dynamic>>[];
  for (final item in raw) {
    if (item is Map) {
      out.add(Map<String, dynamic>.from(item));
    }
  }
  return out;
}

String lineLabel(Map<String, dynamic> line) {
  final pid = line['product_id']?.toString() ?? '?';
  final qty = line['qty'] ?? '?';
  final price = line['unit_price'] ?? line['unitPrice'] ?? '';
  final total = line['line_total'] ?? '';
  return '$pid × $qty'
      '${price.toString().isNotEmpty ? ' @ $price' : ''}'
      '${total.toString().isNotEmpty ? ' = $total' : ''}';
}

/// Default per-line picks: client-only → mine; server-only → server; both → mine if differ.
Map<String, bool> defaultLinePicks({
  required dynamic serverLines,
  required dynamic clientLines,
}) {
  final server = asLineMaps(serverLines);
  final client = asLineMaps(clientLines);
  final serverByKey = {for (final l in server) lineKey(l): l};
  final clientByKey = {for (final l in client) lineKey(l): l};
  final picks = <String, bool>{};
  for (final key in {...serverByKey.keys, ...clientByKey.keys}) {
    final s = serverByKey[key];
    final c = clientByKey[key];
    if (s == null) {
      picks[key] = true;
    } else if (c == null) {
      picks[key] = false;
    } else {
      picks[key] = jsonEncode(s) != jsonEncode(c);
    }
  }
  return picks;
}
