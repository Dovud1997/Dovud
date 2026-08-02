import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/branding/data/branding_repository.dart';

final tenantBrandingProvider = FutureProvider.autoDispose<Map<String, dynamic>>((ref) {
  return ref.watch(brandingRepositoryProvider).fetchTenant();
});

class BrandingPage extends ConsumerStatefulWidget {
  const BrandingPage({super.key});

  @override
  ConsumerState<BrandingPage> createState() => _BrandingPageState();
}

class _BrandingPageState extends ConsumerState<BrandingPage> {
  final _appName = TextEditingController();
  final _logoUrl = TextEditingController();
  final _primary = TextEditingController();
  final _secondary = TextEditingController();
  final _accent = TextEditingController();
  String _themeMode = 'light';
  bool _loaded = false;
  bool _busy = false;
  String? _message;

  @override
  void dispose() {
    _appName.dispose();
    _logoUrl.dispose();
    _primary.dispose();
    _secondary.dispose();
    _accent.dispose();
    super.dispose();
  }

  void _hydrate(Map<String, dynamic> data) {
    if (_loaded) return;
    _appName.text = data['app_name']?.toString() ?? '';
    _logoUrl.text = data['logo_url']?.toString() ?? '';
    _primary.text = data['primary_color']?.toString() ?? '#0F766E';
    _secondary.text = data['secondary_color']?.toString() ?? '#134E4A';
    _accent.text = data['accent_color']?.toString() ?? '#F59E0B';
    _themeMode = data['theme_mode_default']?.toString() ?? 'light';
    _loaded = true;
  }

  Future<void> _save() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      await ref.read(brandingRepositoryProvider).updateTenant({
        'app_name': _appName.text.trim(),
        'logo_url': _logoUrl.text.trim().isEmpty ? null : _logoUrl.text.trim(),
        'primary_color': _primary.text.trim(),
        'secondary_color': _secondary.text.trim(),
        'accent_color': _accent.text.trim(),
        'theme_mode_default': _themeMode,
      });
      await ref.read(sessionControllerProvider.notifier).refreshBranding();
      ref.invalidate(tenantBrandingProvider);
      setState(() {
        _loaded = false;
        _message = 'Branding saved';
      });
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(tenantBrandingProvider);
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text('Branding studio', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text('App name, colors, and logo URL for this tenant.'),
        const SizedBox(height: 16),
        async.when(
          data: (data) {
            _hydrate(data);
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                TextField(controller: _appName, decoration: const InputDecoration(labelText: 'App name')),
                TextField(controller: _logoUrl, decoration: const InputDecoration(labelText: 'Logo URL')),
                TextField(controller: _primary, decoration: const InputDecoration(labelText: 'Primary color')),
                TextField(controller: _secondary, decoration: const InputDecoration(labelText: 'Secondary color')),
                TextField(controller: _accent, decoration: const InputDecoration(labelText: 'Accent color')),
                DropdownButtonFormField<String>(
                  value: _themeMode,
                  decoration: const InputDecoration(labelText: 'Default theme'),
                  items: const [
                    DropdownMenuItem(value: 'light', child: Text('light')),
                    DropdownMenuItem(value: 'dark', child: Text('dark')),
                    DropdownMenuItem(value: 'system', child: Text('system')),
                  ],
                  onChanged: (v) => setState(() => _themeMode = v ?? 'light'),
                ),
                const SizedBox(height: 16),
                FilledButton(onPressed: _busy ? null : _save, child: const Text('Save branding')),
                if (_message != null) ...[
                  const SizedBox(height: 8),
                  Text(_message!),
                ],
                const SizedBox(height: 24),
                Text('Preview', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 8),
                Wrap(
                  spacing: 12,
                  children: [
                    _swatch(_primary.text, 'Primary'),
                    _swatch(_secondary.text, 'Secondary'),
                    _swatch(_accent.text, 'Accent'),
                  ],
                ),
              ],
            );
          },
          loading: () => const LinearProgressIndicator(),
          error: (e, _) => Text('$e'),
        ),
      ],
    );
  }

  Widget _swatch(String hex, String label) {
    Color color;
    try {
      var v = hex.replaceAll('#', '');
      if (v.length == 6) v = 'FF$v';
      color = Color(int.parse(v, radix: 16));
    } catch (_) {
      color = Colors.grey;
    }
    return Column(
      children: [
        Container(width: 56, height: 56, color: color),
        const SizedBox(height: 4),
        Text(label),
      ],
    );
  }
}
