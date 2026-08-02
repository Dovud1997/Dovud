import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/identity/data/identity_repository.dart';

final usersListProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(identityRepositoryProvider).listUsers();
});

final rolesForUsersProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(identityRepositoryProvider).listRoles();
});

class UsersPage extends ConsumerStatefulWidget {
  const UsersPage({super.key});

  @override
  ConsumerState<UsersPage> createState() => _UsersPageState();
}

class _UsersPageState extends ConsumerState<UsersPage> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _fullName = TextEditingController();
  final _phone = TextEditingController();
  String _locale = 'ru';
  final Set<String> _roleIds = {};
  String? _editId;
  String _status = 'active';
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    _fullName.dispose();
    _phone.dispose();
    super.dispose();
  }

  void _resetForm() {
    _editId = null;
    _email.clear();
    _password.clear();
    _fullName.clear();
    _phone.clear();
    _locale = 'ru';
    _status = 'active';
    _roleIds.clear();
  }

  void _loadUser(Map<String, dynamic> u) {
    setState(() {
      _editId = u['id']?.toString();
      _email.text = u['email']?.toString() ?? '';
      _password.clear();
      _fullName.text = u['full_name']?.toString() ?? '';
      _phone.text = u['phone']?.toString() ?? '';
      _locale = u['locale']?.toString() ?? 'ru';
      _status = u['status']?.toString() ?? 'active';
      _roleIds
        ..clear()
        ..addAll((u['role_ids'] as List<dynamic>? ?? const []).map((e) => e.toString()));
    });
  }

  Future<void> _save() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      final repo = ref.read(identityRepositoryProvider);
      if (_editId == null) {
        final created = await repo.createUser({
          'email': _email.text.trim(),
          'password': _password.text,
          'full_name': _fullName.text.trim(),
          if (_phone.text.trim().isNotEmpty) 'phone': _phone.text.trim(),
          'locale': _locale,
          'role_ids': _roleIds.toList(),
        });
        setState(() => _message = 'Created ${created['email']}');
      } else {
        await repo.updateUser(_editId!, {
          'full_name': _fullName.text.trim(),
          'phone': _phone.text.trim(),
          'locale': _locale,
          'status': _status,
        });
        await repo.assignRoles(_editId!, _roleIds.toList());
        setState(() => _message = 'Updated user');
      }
      _resetForm();
      ref.invalidate(usersListProvider);
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final usersAsync = ref.watch(usersListProvider);
    final rolesAsync = ref.watch(rolesForUsersProvider);

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Users', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('Create and manage tenant users, status, and roles.'),
        const SizedBox(height: 16),
        if (_editId != null)
          Text('Editing $_editId', style: Theme.of(context).textTheme.titleMedium)
        else
          Text('New user', style: Theme.of(context).textTheme.titleMedium),
        TextField(
          controller: _email,
          enabled: _editId == null,
          decoration: const InputDecoration(labelText: 'Email'),
        ),
        if (_editId == null)
          TextField(
            controller: _password,
            obscureText: true,
            decoration: const InputDecoration(labelText: 'Password'),
          ),
        TextField(controller: _fullName, decoration: const InputDecoration(labelText: 'Full name')),
        TextField(controller: _phone, decoration: const InputDecoration(labelText: 'Phone')),
        DropdownButtonFormField<String>(
          value: _locale,
          decoration: const InputDecoration(labelText: 'Locale'),
          items: const [
            DropdownMenuItem(value: 'ru', child: Text('ru')),
            DropdownMenuItem(value: 'uz', child: Text('uz')),
            DropdownMenuItem(value: 'en', child: Text('en')),
          ],
          onChanged: (v) => setState(() => _locale = v ?? 'ru'),
        ),
        if (_editId != null)
          DropdownButtonFormField<String>(
            value: _status,
            decoration: const InputDecoration(labelText: 'Status'),
            items: const [
              DropdownMenuItem(value: 'active', child: Text('active')),
              DropdownMenuItem(value: 'disabled', child: Text('disabled')),
              DropdownMenuItem(value: 'invited', child: Text('invited')),
            ],
            onChanged: (v) => setState(() => _status = v ?? 'active'),
          ),
        const SizedBox(height: 8),
        Text('Roles', style: Theme.of(context).textTheme.titleSmall),
        rolesAsync.when(
          data: (roles) => Wrap(
            spacing: 8,
            children: [
              for (final r in roles)
                FilterChip(
                  label: Text(r['code']?.toString() ?? ''),
                  selected: _roleIds.contains(r['id']?.toString()),
                  onSelected: (sel) {
                    setState(() {
                      final id = r['id']?.toString() ?? '';
                      if (sel) {
                        _roleIds.add(id);
                      } else {
                        _roleIds.remove(id);
                      }
                    });
                  },
                ),
            ],
          ),
          loading: () => const LinearProgressIndicator(),
          error: (e, _) => Text('$e'),
        ),
        const SizedBox(height: 12),
        Wrap(
          spacing: 8,
          children: [
            FilledButton(onPressed: _busy ? null : _save, child: Text(_editId == null ? 'Create' : 'Save')),
            if (_editId != null)
              TextButton(
                onPressed: _busy
                    ? null
                    : () => setState(_resetForm),
                child: const Text('Cancel'),
              ),
          ],
        ),
        if (_message != null) ...[
          const SizedBox(height: 8),
          Text(_message!),
        ],
        const SizedBox(height: 24),
        Text('Directory', style: Theme.of(context).textTheme.titleMedium),
        usersAsync.when(
          data: (rows) => rows.isEmpty
              ? const Text('No users')
              : Column(
                  children: [
                    for (final u in rows)
                      ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(u['full_name']?.toString() ?? u['email']?.toString() ?? ''),
                        subtitle: Text(
                          '${u['email']} · ${u['status']} · ${(u['roles'] as List<dynamic>? ?? const []).join(', ')}',
                        ),
                        onTap: () => _loadUser(u),
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
