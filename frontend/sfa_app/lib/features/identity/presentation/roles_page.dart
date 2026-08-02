import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/identity/data/identity_repository.dart';

final rolesListProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(identityRepositoryProvider).listRoles();
});

final permissionsListProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(identityRepositoryProvider).listPermissions();
});

class RolesPage extends ConsumerStatefulWidget {
  const RolesPage({super.key});

  @override
  ConsumerState<RolesPage> createState() => _RolesPageState();
}

class _RolesPageState extends ConsumerState<RolesPage> {
  final _code = TextEditingController();
  final _name = TextEditingController();
  String? _editId;
  bool _isSystem = false;
  final Set<String> _permCodes = {};
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _code.dispose();
    _name.dispose();
    super.dispose();
  }

  void _reset() {
    _editId = null;
    _isSystem = false;
    _code.clear();
    _name.clear();
    _permCodes.clear();
  }

  void _loadRole(Map<String, dynamic> r) {
    setState(() {
      _editId = r['id']?.toString();
      _isSystem = r['is_system'] == true;
      _code.text = r['code']?.toString() ?? '';
      _name.text = r['name']?.toString() ?? '';
      _permCodes
        ..clear()
        ..addAll((r['permission_codes'] as List<dynamic>? ?? const []).map((e) => e.toString()));
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
        final created = await repo.createRole({
          'code': _code.text.trim(),
          'name': _name.text.trim(),
          'permission_codes': _permCodes.toList(),
        });
        setState(() => _message = 'Created ${created['code']}');
      } else {
        if (_isSystem) {
          setState(() => _message = 'System roles are immutable');
          return;
        }
        await repo.setRolePermissions(_editId!, _permCodes.toList());
        setState(() => _message = 'Permissions updated');
      }
      _reset();
      ref.invalidate(rolesListProvider);
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final rolesAsync = ref.watch(rolesListProvider);
    final permsAsync = ref.watch(permissionsListProvider);

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Roles', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('Custom roles and permission matrices. System roles are locked.'),
        const SizedBox(height: 16),
        Text(
          _editId == null ? 'New role' : (_isSystem ? 'System role (read-only)' : 'Edit permissions'),
          style: Theme.of(context).textTheme.titleMedium,
        ),
        TextField(
          controller: _code,
          enabled: _editId == null,
          decoration: const InputDecoration(labelText: 'Code'),
        ),
        TextField(
          controller: _name,
          enabled: _editId == null,
          decoration: const InputDecoration(labelText: 'Name'),
        ),
        const SizedBox(height: 8),
        Text('Permissions', style: Theme.of(context).textTheme.titleSmall),
        permsAsync.when(
          data: (perms) => Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              for (final p in perms)
                FilterChip(
                  label: Text(p['code']?.toString() ?? ''),
                  selected: _permCodes.contains(p['code']?.toString()),
                  onSelected: (_editId != null && _isSystem)
                      ? null
                      : (sel) {
                          setState(() {
                            final code = p['code']?.toString() ?? '';
                            if (sel) {
                              _permCodes.add(code);
                            } else {
                              _permCodes.remove(code);
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
            FilledButton(
              onPressed: _busy || (_editId != null && _isSystem) ? null : _save,
              child: Text(_editId == null ? 'Create' : 'Save permissions'),
            ),
            if (_editId != null)
              TextButton(onPressed: _busy ? null : () => setState(_reset), child: const Text('Cancel')),
          ],
        ),
        if (_message != null) ...[
          const SizedBox(height: 8),
          Text(_message!),
        ],
        const SizedBox(height: 24),
        Text('All roles', style: Theme.of(context).textTheme.titleMedium),
        rolesAsync.when(
          data: (rows) => Column(
            children: [
              for (final r in rows)
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text('${r['name']} (${r['code']})'),
                  subtitle: Text(
                    r['is_system'] == true
                        ? 'system · ${(r['permission_codes'] as List<dynamic>? ?? const []).length} perms'
                        : 'custom · ${(r['permission_codes'] as List<dynamic>? ?? const []).length} perms',
                  ),
                  onTap: () => _loadRole(r),
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
