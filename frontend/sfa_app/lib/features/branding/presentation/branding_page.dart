import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';
import 'package:sfa_app/features/branding/data/branding_repository.dart';
import 'package:sfa_app/features/documents/data/documents_repository.dart';

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

  Future<void> _uploadLogo() async {
    setState(() {
      _busy = true;
      _message = null;
    });
    try {
      final picked = await FilePicker.platform.pickFiles(
        type: FileType.image,
        withData: true,
      );
      if (picked == null || picked.files.isEmpty) {
        setState(() => _busy = false);
        return;
      }
      final file = picked.files.first;
      final bytes = file.bytes;
      if (bytes == null || bytes.isEmpty) {
        throw StateError('Could not read file bytes');
      }
      final mime = _guessMime(file.name, file.extension);
      final uploaded = await ref.read(documentsRepositoryProvider).uploadBytes(
            fileName: file.name,
            mime: mime,
            bytes: bytes,
          );
      final branding = await ref.read(brandingRepositoryProvider).attachAsset(
            fileId: uploaded['id'].toString(),
            kind: 'logo',
          );
      _logoUrl.text = branding['logo_url']?.toString() ?? _logoUrl.text;
      await ref.read(sessionControllerProvider.notifier).refreshBranding();
      ref.invalidate(tenantBrandingProvider);
      setState(() {
        _loaded = false;
        _message = 'Logo uploaded';
      });
    } catch (e) {
      setState(() => _message = '$e');
    } finally {
      setState(() => _busy = false);
    }
  }

  String _guessMime(String name, String? ext) {
    final e = (ext ?? name.split('.').last).toLowerCase();
    switch (e) {
      case 'jpg':
      case 'jpeg':
        return 'image/jpeg';
      case 'webp':
        return 'image/webp';
      case 'gif':
        return 'image/gif';
      case 'svg':
        return 'image/svg+xml';
      default:
        return 'image/png';
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
        Text('App name, colors, and logo (URL or upload via presign).'),
        const SizedBox(height: 16),
        async.when(
          data: (data) {
            _hydrate(data);
            final logo = _logoUrl.text.trim();
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                TextField(controller: _appName, decoration: const InputDecoration(labelText: 'App name')),
                TextField(controller: _logoUrl, decoration: const InputDecoration(labelText: 'Logo URL')),
                const SizedBox(height: 8),
                FilledButton.tonal(
                  onPressed: _busy ? null : _uploadLogo,
                  child: const Text('Upload logo'),
                ),
                if (logo.isNotEmpty) ...[
                  const SizedBox(height: 12),
                  ClipRRect(
                    borderRadius: BorderRadius.circular(8),
                    child: Image.network(
                      logo,
                      height: 72,
                      errorBuilder: (_, __, ___) => const Text('Logo preview unavailable'),
                    ),
                  ),
                ],
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
