import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/tenant/data/domains_repository.dart';

final domainsListProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(domainsRepositoryProvider).list();
});

class DomainsPage extends ConsumerStatefulWidget {
  const DomainsPage({super.key});

  @override
  ConsumerState<DomainsPage> createState() => _DomainsPageState();
}

class _DomainsPageState extends ConsumerState<DomainsPage> {
  final _host = TextEditingController();
  bool _isPrimary = false;
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _host.dispose();
    super.dispose();
  }

  Future<void> _add() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      await ref.read(domainsRepositoryProvider).add(
            host: _host.text.trim(),
            isPrimary: _isPrimary,
          );
      _host.clear();
      _isPrimary = false;
      ref.invalidate(domainsListProvider);
      setState(() => _message = 'Domain added');
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  Future<void> _delete(String id) async {
    setState(() => _busy = true);
    try {
      await ref.read(domainsRepositoryProvider).delete(id);
      ref.invalidate(domainsListProvider);
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(domainsListProvider);
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Custom domains', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('Hosts used for public branding resolution (?host=).'),
        const SizedBox(height: 16),
        TextField(
          controller: _host,
          decoration: const InputDecoration(
            labelText: 'Host',
            hintText: 'app.example.com',
          ),
        ),
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          title: const Text('Primary domain'),
          value: _isPrimary,
          onChanged: (v) => setState(() => _isPrimary = v),
        ),
        FilledButton(onPressed: _busy ? null : _add, child: const Text('Add domain')),
        if (_message != null) ...[
          const SizedBox(height: 8),
          Text(_message!),
        ],
        const SizedBox(height: 24),
        Text('Configured', style: Theme.of(context).textTheme.titleMedium),
        async.when(
          data: (rows) => rows.isEmpty
              ? const Text('No domains yet')
              : Column(
                  children: [
                    for (final d in rows)
                      ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(d['host']?.toString() ?? ''),
                        subtitle: Text(d['is_primary'] == true ? 'primary' : 'alias'),
                        trailing: IconButton(
                          icon: const Icon(Icons.delete_outline),
                          onPressed: _busy ? null : () => _delete(d['id'].toString()),
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
