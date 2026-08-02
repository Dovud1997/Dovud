import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/network/api_client.dart';

final portalLinksProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) async {
  final envelope = await ref.watch(apiClientProvider).get('/portal/links');
  final data = envelope['data'] as List<dynamic>? ?? const [];
  return data.map((e) => Map<String, dynamic>.from(e as Map)).toList();
});

class PortalLinksPage extends ConsumerStatefulWidget {
  const PortalLinksPage({super.key});

  @override
  ConsumerState<PortalLinksPage> createState() => _PortalLinksPageState();
}

class _PortalLinksPageState extends ConsumerState<PortalLinksPage> {
  final _userId = TextEditingController();
  final _customerId = TextEditingController();
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _userId.dispose();
    _customerId.dispose();
    super.dispose();
  }

  Future<void> _link() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      await ref.read(apiClientProvider).post('/portal/links', data: {
        'user_id': _userId.text.trim(),
        'customer_id': _customerId.text.trim(),
      });
      ref.invalidate(portalLinksProvider);
      setState(() => _message = 'Linked');
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  Future<void> _unlink(String userId) async {
    setState(() => _busy = true);
    try {
      await ref.read(apiClientProvider).delete('/portal/links/$userId');
      ref.invalidate(portalLinksProvider);
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(portalLinksProvider);
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Portal user links', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('Attach a portal login to a CRM customer.'),
        const SizedBox(height: 16),
        TextField(controller: _userId, decoration: const InputDecoration(labelText: 'User ID')),
        TextField(controller: _customerId, decoration: const InputDecoration(labelText: 'Customer ID')),
        const SizedBox(height: 12),
        FilledButton(onPressed: _busy ? null : _link, child: const Text('Link user')),
        if (_message != null) ...[
          const SizedBox(height: 8),
          Text(_message!),
        ],
        const SizedBox(height: 24),
        async.when(
          data: (rows) => rows.isEmpty
              ? const Text('No links yet')
              : Column(
                  children: [
                    for (final r in rows)
                      ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text('User ${r['user_id']}'),
                        subtitle: Text('Customer ${r['customer_id']}'),
                        trailing: IconButton(
                          icon: const Icon(Icons.link_off),
                          onPressed: _busy ? null : () => _unlink(r['user_id'].toString()),
                        ),
                      ),
                  ],
                ),
          loading: () => const LinearProgressIndicator(),
          error: (e, _) => Text('$e'),
        ),
      ],
    );
  }
}
