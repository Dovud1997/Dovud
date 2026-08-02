import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/features/sync/presentation/conflict_line_merge.dart';

void main() {
  test('mergeOrderLines picks client/server per product', () {
    final server = [
      {'product_id': 'a', 'qty': 1, 'unit_price': 10},
      {'product_id': 'b', 'qty': 2, 'unit_price': 20},
    ];
    final client = [
      {'product_id': 'a', 'qty': 5, 'unit_price': 10},
      {'product_id': 'c', 'qty': 1, 'unit_price': 30},
    ];
    final picks = defaultLinePicks(serverLines: server, clientLines: client);
    expect(picks['p:a'], isTrue); // both differ → mine
    expect(picks['p:b'], isFalse); // server-only
    expect(picks['p:c'], isTrue); // client-only

    final merged = mergeOrderLines(
      serverLines: server,
      clientLines: client,
      pickClient: picks,
    );
    expect(merged.map((e) => e['product_id']).toList(), ['a', 'b', 'c']);
    expect(merged.firstWhere((e) => e['product_id'] == 'a')['qty'], 5);
    expect(merged.firstWhere((e) => e['product_id'] == 'b')['qty'], 2);
    expect(merged.firstWhere((e) => e['product_id'] == 'c')['qty'], 1);

    // Drop client-only by flipping pick
    picks['p:c'] = false;
    final withoutC = mergeOrderLines(
      serverLines: server,
      clientLines: client,
      pickClient: picks,
    );
    expect(withoutC.any((e) => e['product_id'] == 'c'), isFalse);
  });
}
