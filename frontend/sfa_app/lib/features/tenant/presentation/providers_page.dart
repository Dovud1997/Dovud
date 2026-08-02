import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/tenant/data/providers_repository.dart';

final providersListProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(providersRepositoryProvider).list();
});

class ProvidersPage extends ConsumerStatefulWidget {
  const ProvidersPage({super.key});

  @override
  ConsumerState<ProvidersPage> createState() => _ProvidersPageState();
}

class _ProvidersPageState extends ConsumerState<ProvidersPage> {
  String _type = 'smtp';
  String _driver = 'log';
  bool _enabled = true;
  final _from = TextEditingController();
  final _host = TextEditingController(text: 'localhost');
  final _port = TextEditingController(text: '1025');
  final _username = TextEditingController();
  final _password = TextEditingController();
  final _webhook = TextEditingController();
  final _projectId = TextEditingController();
  final _serviceAccount = TextEditingController();
  final _testTo = TextEditingController();
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _from.dispose();
    _host.dispose();
    _port.dispose();
    _username.dispose();
    _password.dispose();
    _webhook.dispose();
    _projectId.dispose();
    _serviceAccount.dispose();
    _testTo.dispose();
    super.dispose();
  }

  void _selectType(String type, List<Map<String, dynamic>> rows) {
    setState(() {
      _type = type;
      _password.clear();
      _serviceAccount.clear();
      final match = rows.where((e) => e['type'] == type).toList();
      if (match.isEmpty) {
        _driver = 'log';
        _enabled = true;
        return;
      }
      final row = match.first;
      final cfg = Map<String, dynamic>.from(row['config'] as Map? ?? const {});
      final drivers = _driversForType(type);
      final d = row['driver']?.toString() ?? 'log';
      _driver = drivers.contains(d) ? d : 'log';
      _enabled = row['is_enabled'] == true;
      _from.text = cfg['from']?.toString() ?? '';
      _host.text = cfg['host']?.toString() ?? 'localhost';
      _port.text = '${cfg['port'] ?? 1025}';
      _username.text = cfg['username']?.toString() ?? '';
      _webhook.text = cfg['webhook_url']?.toString() ?? '';
      _projectId.text = cfg['project_id']?.toString() ?? '';
      final sa = cfg['service_account_json']?.toString() ?? '';
      _serviceAccount.text = sa == '********' ? '' : sa;
    });
  }

  Future<void> _save() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      final config = <String, dynamic>{};
      if (_type == 'smtp') {
        config.addAll({
          'from': _from.text.trim(),
          'host': _host.text.trim(),
          'port': int.tryParse(_port.text.trim()) ?? 1025,
          'username': _username.text.trim(),
          if (_password.text.isNotEmpty) 'password': _password.text,
        });
      } else if (_type == 'push' && _driver == 'fcm') {
        config['project_id'] = _projectId.text.trim();
        if (_serviceAccount.text.trim().isNotEmpty) {
          config['service_account_json'] = _serviceAccount.text.trim();
        }
      } else {
        config['webhook_url'] = _webhook.text.trim();
      }
      await ref.read(providersRepositoryProvider).upsert(_type, {
        'driver': _driver,
        'is_enabled': _enabled,
        'config': config,
      });
      ref.invalidate(providersListProvider);
      setState(() => _message = 'Saved $_type provider');
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  Future<void> _test() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      await ref.read(providersRepositoryProvider).test(_type, to: _testTo.text.trim());
      setState(() => _message = 'Test sent via $_type');
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(providersListProvider);
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Tenant providers', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text(
          'SMTP / SMS / Push credentials (secrets encrypted at rest). Push supports log, http webhook, FCM.',
          style: Theme.of(context).textTheme.bodyMedium,
        ),
        const SizedBox(height: 16),
        async.when(
          data: (rows) => Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Wrap(
                spacing: 8,
                children: [
                  for (final t in const ['smtp', 'sms', 'push'])
                    ChoiceChip(
                      label: Text(t.toUpperCase()),
                      selected: _type == t,
                      onSelected: (_) => _selectType(t, rows),
                    ),
                ],
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String>(
                value: _driver,
                decoration: const InputDecoration(labelText: 'Driver'),
                items: [
                  for (final d in _driversForType(_type))
                    DropdownMenuItem(value: d, child: Text(d)),
                ],
                onChanged: (v) => setState(() => _driver = v ?? 'log'),
              ),
              SwitchListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Enabled'),
                value: _enabled,
                onChanged: (v) => setState(() => _enabled = v),
              ),
              if (_type == 'smtp') ...[
                TextField(controller: _from, decoration: const InputDecoration(labelText: 'From')),
                TextField(controller: _host, decoration: const InputDecoration(labelText: 'Host')),
                TextField(controller: _port, decoration: const InputDecoration(labelText: 'Port')),
                TextField(controller: _username, decoration: const InputDecoration(labelText: 'Username')),
                TextField(
                  controller: _password,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: 'Password (leave blank to keep)'),
                ),
              ] else if (_type == 'push' && _driver == 'fcm') ...[
                TextField(
                  controller: _projectId,
                  decoration: const InputDecoration(labelText: 'FCM project ID'),
                ),
                TextField(
                  controller: _serviceAccount,
                  maxLines: 6,
                  decoration: const InputDecoration(
                    labelText: 'Service account JSON',
                    hintText: 'Paste Firebase service account JSON (leave blank to keep)',
                  ),
                ),
              ] else
                TextField(
                  controller: _webhook,
                  decoration: const InputDecoration(labelText: 'Webhook URL'),
                ),
              TextField(
                controller: _testTo,
                decoration: InputDecoration(
                  labelText: _type == 'push' ? 'Test device token' : 'Test recipient',
                ),
              ),
              const SizedBox(height: 16),
              Wrap(
                spacing: 8,
                children: [
                  FilledButton(onPressed: _busy ? null : _save, child: const Text('Save')),
                  FilledButton.tonal(onPressed: _busy ? null : _test, child: const Text('Send test')),
                ],
              ),
              if (_message != null) ...[
                const SizedBox(height: 12),
                Text(_message!),
              ],
              const SizedBox(height: 24),
              Text('Configured', style: Theme.of(context).textTheme.titleMedium),
              if (rows.isEmpty) const Text('No providers saved yet'),
              for (final r in rows)
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text('${r['type']} · ${r['driver']}'),
                  subtitle: Text(r['is_enabled'] == true ? 'enabled' : 'disabled'),
                  onTap: () => _selectType(r['type'].toString(), rows),
                ),
            ],
          ),
          loading: () => const LinearProgressIndicator(),
          error: (e, _) => Text('$e'),
        ),
      ],
    );
  }

  List<String> _driversForType(String type) {
    if (type == 'smtp') return const ['log', 'file', 'smtp'];
    if (type == 'push') return const ['log', 'http', 'fcm'];
    return const ['log', 'http'];
  }
}
