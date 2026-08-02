import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';

class ConflictResolvePage extends ConsumerStatefulWidget {
  const ConflictResolvePage({super.key, required this.conflictId});

  final String conflictId;

  @override
  ConsumerState<ConflictResolvePage> createState() => _ConflictResolvePageState();
}

class _ConflictResolvePageState extends ConsumerState<ConflictResolvePage> {
  SyncConflict? _conflict;
  String? _error;
  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final rows = await ref.read(syncRepositoryProvider).listConflicts();
      final match = rows.where((c) => c.id == widget.conflictId).toList();
      if (match.isEmpty) {
        setState(() => _error = 'Conflict not found or already resolved');
      } else {
        setState(() => _conflict = match.first);
      }
    } catch (e) {
      setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _resolve(String resolution) async {
    setState(() => _busy = true);
    try {
      final repo = ref.read(syncRepositoryProvider);
      final resolved = await repo.resolveConflict(
        conflictId: widget.conflictId,
        resolution: resolution,
      );
      final opId = resolved.clientOpId;
      if (opId.isNotEmpty) {
        await ref.read(localOutboxProvider).removeByOpIds([opId]);
      }
      await ref.read(offlineStoreProvider).pullAndCache();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Resolved: $resolution')),
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

  String _pretty(Map<String, dynamic> m) {
    try {
      return const JsonEncoder.withIndent('  ').convert(m);
    } catch (_) {
      return m.toString();
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = _conflict;
    return Scaffold(
      appBar: AppBar(title: const Text('Resolve conflict')),
      body: _busy && c == null
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text(_error!))
              : c == null
                  ? const SizedBox.shrink()
                  : ListView(
                      padding: const EdgeInsets.all(24),
                      children: [
                        Text('${c.entityType} · ${c.entityId}',
                            style: Theme.of(context).textTheme.titleMedium),
                        Text('base v${c.baseVersion} → server v${c.serverVersion}'),
                        const SizedBox(height: 16),
                        Text('Server', style: Theme.of(context).textTheme.titleSmall),
                        SelectableText(_pretty(c.serverPayload)),
                        const SizedBox(height: 16),
                        Text('Yours', style: Theme.of(context).textTheme.titleSmall),
                        SelectableText(_pretty(c.clientPayload)),
                        const SizedBox(height: 24),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: [
                            FilledButton(
                              onPressed: _busy ? null : () => _resolve('server_wins'),
                              child: const Text('Take server'),
                            ),
                            FilledButton.tonal(
                              onPressed: _busy ? null : () => _resolve('client_wins'),
                              child: const Text('Keep mine'),
                            ),
                          ],
                        ),
                      ],
                    ),
    );
  }
}
