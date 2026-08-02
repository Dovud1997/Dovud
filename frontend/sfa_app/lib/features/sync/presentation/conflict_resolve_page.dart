import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/local_outbox.dart';
import 'package:sfa_app/core/offline/offline_store.dart';
import 'package:sfa_app/features/sync/data/sync_repository.dart';
import 'package:sfa_app/features/sync/presentation/conflict_line_merge.dart';

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
  /// true = keep client value for that field (excluding `lines`)
  final Map<String, bool> _pickClient = {};
  /// true = keep client line for that line key
  final Map<String, bool> _pickLineClient = {};
  bool _linesDiffer = false;

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
          ..addEntries(
            diffs.where((k) => k != 'lines').map((k) => MapEntry(k, true)),
          );
        _linesDiffer = diffs.contains('lines') &&
            (c.serverPayload['lines'] is List || c.clientPayload['lines'] is List);
        _pickLineClient
          ..clear()
          ..addAll(
            _linesDiffer
                ? defaultLinePicks(
                    serverLines: c.serverPayload['lines'],
                    clientLines: c.clientPayload['lines'],
                  )
                : const {},
          );
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
    if (_linesDiffer) {
      out['lines'] = mergeOrderLines(
        serverLines: c.serverPayload['lines'],
        clientLines: c.clientPayload['lines'],
        pickClient: _pickLineClient,
      );
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

  List<Widget> _lineMergeTiles(SyncConflict c) {
    if (!_linesDiffer) return const [];
    final server = {for (final l in asLineMaps(c.serverPayload['lines'])) lineKey(l): l};
    final client = {for (final l in asLineMaps(c.clientPayload['lines'])) lineKey(l): l};
    final keys = _pickLineClient.keys.toList()..sort();
    return [
      const SizedBox(height: 8),
      Text('Order lines', style: Theme.of(context).textTheme.titleSmall),
      Text(
        'Toggle per product: on = yours, off = server',
        style: Theme.of(context).textTheme.bodySmall,
      ),
      ...keys.map((key) {
        final pickMine = _pickLineClient[key] ?? true;
        final s = server[key];
        final cl = client[key];
        return CheckboxListTile(
          contentPadding: EdgeInsets.zero,
          value: pickMine,
          onChanged: _busy ? null : (v) => setState(() => _pickLineClient[key] = v ?? true),
          title: Text(cl != null ? lineLabel(cl) : (s != null ? lineLabel(s) : key)),
          subtitle: Text(
            [
              if (s != null) 'server: ${lineLabel(s)}',
              if (cl != null) 'yours: ${lineLabel(cl)}',
              if (s == null) 'server: (none)',
              if (cl == null) 'yours: (none)',
            ].join('\n'),
          ),
          secondary: Text(pickMine ? 'Yours' : 'Server'),
        );
      }),
    ];
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
                        if (_pickClient.isEmpty && !_linesDiffer)
                          const Text('No field differences — use Take server / Keep mine')
                        else ...[
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
                                'server: ${_pretty(c.serverPayload[key])}\n'
                                'yours: ${_pretty(c.clientPayload[key])}',
                              ),
                              secondary: Text(pickMine ? 'Yours' : 'Server'),
                            );
                          }),
                          ..._lineMergeTiles(c),
                        ],
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
