import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/agent_write_coordinator.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/features/catalog/data/catalog_repository.dart';
import 'package:sfa_app/features/crm/data/crm_repository.dart';
import 'package:sfa_app/features/returns/data/returns_repository.dart';

class _LineDraft {
  _LineDraft({
    required this.productId,
    required this.name,
    required this.sku,
    required this.unitPrice,
  });

  final String productId;
  final String name;
  final String sku;
  double unitPrice;
  double qty = 1;

  double get lineTotal => qty * unitPrice;
}

class ReturnComposePage extends ConsumerStatefulWidget {
  const ReturnComposePage({super.key});

  @override
  ConsumerState<ReturnComposePage> createState() => _ReturnComposePageState();
}

class _ReturnComposePageState extends ConsumerState<ReturnComposePage> {
  bool _busy = false;
  String? _error;
  List<Map<String, dynamic>> _customers = const [];
  List<Map<String, dynamic>> _products = const [];
  final Map<String, double> _priceByProduct = {};
  String? _customerId;
  final _reasonCtrl = TextEditingController();
  final List<_LineDraft> _lines = [];

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _reasonCtrl.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final store = ref.read(offlineStoreProvider);
      final catalog = ref.read(catalogRepositoryProvider);
      List<Map<String, dynamic>> customers;
      List<Map<String, dynamic>> products;
      try {
        customers = await ref.read(crmRepositoryProvider).listCustomers();
        products = await catalog.listProducts();
        await store.warmEntities('customer', customers);
        await store.warmEntities('product', products);
        final lists = await catalog.listPriceLists();
        if (lists.isNotEmpty) {
          final def = lists.firstWhere(
            (e) => e['is_default'] == true,
            orElse: () => lists.first,
          );
          final prices = await catalog.listPrices(def['id']?.toString() ?? '');
          await store.warmEntities('product_price', prices);
          for (final pr in prices) {
            final pid = pr['product_id']?.toString() ?? '';
            final amount = (pr['amount'] as num?)?.toDouble() ?? 0;
            if (pid.isNotEmpty) _priceByProduct[pid] = amount;
          }
        }
      } catch (_) {
        customers = await store.listEntities('customer');
        products = await store.listEntities('product');
        final prices = await store.listEntities('product_price');
        for (final pr in prices) {
          final pid = pr['product_id']?.toString() ?? '';
          final amount = (pr['amount'] as num?)?.toDouble() ?? 0;
          if (pid.isNotEmpty) _priceByProduct[pid] = amount;
        }
      }
      setState(() {
        _customers = customers;
        _products = products.where((p) => p['is_active'] != false).toList();
        _customerId = customers.isNotEmpty ? customers.first['id']?.toString() : null;
      });
    } catch (e) {
      setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _addProduct() async {
    if (_products.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('No products — sync catalog first')),
      );
      return;
    }
    final picked = await showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      builder: (ctx) {
        return SafeArea(
          child: SizedBox(
            height: MediaQuery.of(ctx).size.height * 0.6,
            child: ListView.separated(
              itemCount: _products.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (_, i) {
                final p = _products[i];
                final id = p['id']?.toString() ?? '';
                final price = _priceByProduct[id] ?? 0;
                return ListTile(
                  title: Text(p['name']?.toString() ?? ''),
                  subtitle: Text('${p['sku'] ?? ''} · ${price.toStringAsFixed(0)}'),
                  onTap: () => Navigator.pop(ctx, p),
                );
              },
            ),
          ),
        );
      },
    );
    if (picked == null) return;
    final id = picked['id']?.toString() ?? '';
    if (id.isEmpty) return;
    final existing = _lines.indexWhere((l) => l.productId == id);
    setState(() {
      if (existing >= 0) {
        _lines[existing].qty += 1;
      } else {
        _lines.add(_LineDraft(
          productId: id,
          name: picked['name']?.toString() ?? '',
          sku: picked['sku']?.toString() ?? '',
          unitPrice: _priceByProduct[id] ?? 0,
        ));
      }
    });
  }

  double get _grandTotal => _lines.fold(0, (s, l) => s + l.lineTotal);

  Future<void> _submit() async {
    final customerId = _customerId;
    if (customerId == null || customerId.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Select a customer')),
      );
      return;
    }
    if (_lines.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Add at least one product')),
      );
      return;
    }
    setState(() => _busy = true);
    try {
      final reason = _reasonCtrl.text.trim();
      final lines = _lines
          .map((l) => {
                'product_id': l.productId,
                'qty': l.qty,
                'unit_price': l.unitPrice,
                'line_total': l.lineTotal,
                if (reason.isNotEmpty) 'reason': reason,
              })
          .toList();
      await ref.read(agentWriteCoordinatorProvider).write(
            entityType: 'return',
            op: 'create',
            payload: {
              'customer_id': customerId,
              'currency': 'UZS',
              'status': 'draft',
              if (reason.isNotEmpty) 'reason': reason,
              'lines': lines,
            },
            online: () => ref.read(returnsRepositoryProvider).createDraft(
                  customerId: customerId,
                  reason: reason.isEmpty ? null : reason,
                  lines: lines,
                ),
          );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Draft return created')),
        );
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('New return')),
      body: _busy && _customers.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!))
              : ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    DropdownButtonFormField<String>(
                      value: _customerId,
                      decoration: const InputDecoration(labelText: 'Customer'),
                      items: _customers
                          .map((c) => DropdownMenuItem(
                                value: c['id']?.toString(),
                                child: Text(c['name']?.toString() ?? c['code']?.toString() ?? ''),
                              ))
                          .toList(),
                      onChanged: _busy ? null : (v) => setState(() => _customerId = v),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _reasonCtrl,
                      decoration: const InputDecoration(labelText: 'Reason (optional)'),
                      maxLines: 2,
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Text('Lines', style: Theme.of(context).textTheme.titleMedium),
                        const Spacer(),
                        TextButton.icon(
                          onPressed: _busy ? null : _addProduct,
                          icon: const Icon(Icons.add),
                          label: const Text('Add product'),
                        ),
                      ],
                    ),
                    if (_lines.isEmpty)
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 24),
                        child: Text('No lines yet'),
                      )
                    else
                      ..._lines.asMap().entries.map((e) {
                        final i = e.key;
                        final line = e.value;
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Expanded(
                                    child: Text('${line.name}\n${line.sku}',
                                        style: Theme.of(context).textTheme.bodyMedium),
                                  ),
                                  IconButton(
                                    onPressed: _busy
                                        ? null
                                        : () => setState(() => _lines.removeAt(i)),
                                    icon: const Icon(Icons.delete_outline),
                                  ),
                                ],
                              ),
                              Row(
                                children: [
                                  Expanded(
                                    child: TextFormField(
                                      initialValue: line.qty.toStringAsFixed(
                                          line.qty == line.qty.roundToDouble() ? 0 : 2),
                                      decoration: const InputDecoration(labelText: 'Qty'),
                                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                                      onChanged: (v) {
                                        final n = double.tryParse(v.replaceAll(',', '.'));
                                        if (n != null && n > 0) {
                                          setState(() => line.qty = n);
                                        }
                                      },
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: TextFormField(
                                      initialValue: line.unitPrice.toStringAsFixed(0),
                                      decoration: const InputDecoration(labelText: 'Unit price'),
                                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                                      onChanged: (v) {
                                        final n = double.tryParse(v.replaceAll(',', '.'));
                                        if (n != null && n >= 0) {
                                          setState(() => line.unitPrice = n);
                                        }
                                      },
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Text(line.lineTotal.toStringAsFixed(0)),
                                ],
                              ),
                              const Divider(),
                            ],
                          ),
                        );
                      }),
                    Text('Total: ${_grandTotal.toStringAsFixed(0)} UZS',
                        style: Theme.of(context).textTheme.titleMedium),
                    const SizedBox(height: 24),
                    FilledButton(
                      onPressed: _busy ? null : _submit,
                      child: _busy
                          ? const SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Text('Create draft return'),
                    ),
                  ],
                ),
    );
  }
}
