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
  /// true = keep client value for that field
  final Map<String, bool> _pickClient = {};

  static const _skipKeys = {'id', 'version', 'created_at', 'updated_at'};

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
        final c = match.first;
        final diffs = _diffKeys(c.serverPayload, c.clientPayload);
        _pickClient
          ..clear()
          ..addEntries(diffs.map((k) => MapEntry(k, true)));
        setState(() => _conflict = c);
      }
    } catch (e) {
      setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  List<String> _diffKeys(Map<String, dynamic> server, Map<String, dynamic> client) {
    final keys = {...server.keys, ...client.keys}
        .where((k) => !_skipKeys.contains(k))
        .toList()
      ..sort();
    return keys.where((k) {
      return jsonEncode(server[k]) != jsonEncode(client[k]);
    }).toList();
  }

  Map<String, dynamic> _mergedPayload(SyncConflict c) {
    final out = Map<String, dynamic>.from(c.serverPayload);
    for (final e in _pickClient.entries) {
      if (e.value) {
        out[e.key] = c.clientPayload[e.key];
      }
    }
    return out;
  }

  Future<void> _resolve(String resolution, {Map<String, dynamic>? merged}) async {
    setState(() => _busy = true);
    try {
      final repo = ref.read(syncRepositoryProvider);
      final resolved = await repo.resolveConflict(
        conflictId: widget.conflictId,
        resolution: resolution,
        mergedPayload: merged,
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

  String _pretty(dynamic v) {
    try {
      return const JsonEncoder.withIndent('  ').convert(v);
    } catch (_) {
      return '$v';
    }
  }

  String _linesSummary(dynamic v) {
    if (v is! List) return _pretty(v);
    if (v.isEmpty) return '0 lines';
    final parts = v.take(5).map((e) {
      if (e is Map) {
        return '${e['product_id'] ?? '?'}×${e['qty'] ?? '?'}';
      }
      return '$e';
    }).toList();
    final more = v.length > 5 ? ' +${v.length - 5} more' : '';
    return '${v.length} lines: ${parts.join(', ')}$more';
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
                        Text('Field merge', style: Theme.of(context).textTheme.titleSmall),
                        if (_pickClient.isEmpty)
                          const Text('No field differences — use Take server / Keep mine')
                        else
                          ..._pickClient.keys.map((key) {
                            final pickMine = _pickClient[key] ?? true;
                            return CheckboxListTile(
                              contentPadding: EdgeInsets.zero,
                              value: pickMine,
                              onChanged: _busy
                                  ? null
                                  : (v) => setState(() => _pickClient[key] = v ?? true),
                              title: Text(key),
                              subtitle: Text(
                                key == 'lines'
                                    ? 'server: ${_linesSummary(c.serverPayload[key])}\n'
                                        'yours: ${_linesSummary(c.clientPayload[key])}'
                                    : 'server: ${_pretty(c.serverPayload[key])}\n'
                                        'yours: ${_pretty(c.clientPayload[key])}',
                              ),
                              secondary: Text(pickMine ? 'Yours' : 'Server'),
                            );
                          }),
                        const SizedBox(height: 12),
                        Wrap(
                          spacing: 8,
                          runSpacing: 8,
                          children: [
                            FilledButton(
                              onPressed: _busy
                                  ? null
                                  : () => _resolve('merge', merged: _mergedPayload(c)),
                              child: const Text('Apply merge'),
                            ),
                            FilledButton.tonal(
                              onPressed: _busy ? null : () => _resolve('server_wins'),
                              child: const Text('Take server'),
                            ),
                            OutlinedButton(
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
