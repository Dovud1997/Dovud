import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/agent_write_coordinator.dart';
import 'package:sfa_app/core/offline/gps_queue.dart';
import 'package:sfa_app/core/offline/sync_worker.dart';
import 'package:sfa_app/features/crm/data/crm_repository.dart';
import 'package:sfa_app/features/fieldforce/data/fieldforce_repository.dart';

final routesProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(fieldForceRepositoryProvider).listRoutes();
});

final visitsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(fieldForceRepositoryProvider).listVisits();
});

class RoutesPage extends ConsumerStatefulWidget {
  const RoutesPage({super.key});

  @override
  ConsumerState<RoutesPage> createState() => _RoutesPageState();
}

class _RoutesPageState extends ConsumerState<RoutesPage> {
  bool _busy = false;
  int _gpsPending = 0;

  @override
  void initState() {
    super.initState();
    _refreshGpsCount();
  }

  Future<void> _refreshGpsCount() async {
    try {
      final n = await ref.read(gpsQueueProvider).pendingCount();
      if (mounted) setState(() => _gpsPending = n);
    } catch (_) {}
  }

  Future<void> _checkIn() async {
    setState(() => _busy = true);
    try {
      final agents = await ref.read(fieldForceRepositoryProvider).listAgents();
      final customers = await ref.read(crmRepositoryProvider).listCustomers();
      if (agents.isEmpty || customers.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Need an agent and a customer for check-in')),
          );
        }
        return;
      }
      final agentId = agents.first['id']?.toString() ?? '';
      final customerId = customers.first['id']?.toString() ?? '';
      if (agentId.isEmpty || customerId.isEmpty) return;

      await ref.read(agentWriteCoordinatorProvider).write(
            entityType: 'visit',
            op: 'create',
            payload: {
              'agent_id': agentId,
              'customer_id': customerId,
              'result': '',
            },
            online: () => ref.read(fieldForceRepositoryProvider).checkIn(
                  agentId: agentId,
                  customerId: customerId,
                ),
          );
      ref.invalidate(visitsProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Checked in')));
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _checkOut(String visitId) async {
    setState(() => _busy = true);
    try {
      await ref.read(fieldForceRepositoryProvider).checkOut(visitId: visitId, result: 'success');
      ref.invalidate(visitsProvider);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _logGps() async {
    setState(() => _busy = true);
    try {
      final agents = await ref.read(fieldForceRepositoryProvider).listAgents();
      if (agents.isEmpty) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('No agent profile')),
          );
        }
        return;
      }
      final agentId = agents.first['id']?.toString() ?? '';
      if (agentId.isEmpty) return;
      // Demo coords (Tashkent) — device GPS can replace later.
      await ref.read(gpsQueueProvider).enqueue(
            agentId: agentId,
            lat: 41.3111,
            lng: 69.2797,
            accuracy: 25,
          );
      await _refreshGpsCount();
      final flush = await ref.read(gpsQueueProvider).flush();
      await _refreshGpsCount();
      if (flush['uploaded'] == 0) {
        await ref.read(syncWorkerProvider).tick(reason: 'gps');
        await _refreshGpsCount();
      }
      if (mounted) {
        final pending = _gpsPending;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              pending > 0
                  ? 'GPS queued ($pending pending — will sync offline)'
                  : 'GPS point uploaded',
            ),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final routes = ref.watch(routesProvider);
    final visits = ref.watch(visitsProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Routes & visits'),
        actions: [
          TextButton.icon(
            onPressed: _busy ? null : _logGps,
            icon: const Icon(Icons.my_location),
            label: Text(_gpsPending > 0 ? 'GPS ($_gpsPending)' : 'GPS'),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _busy ? null : _checkIn,
        icon: _busy
            ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2))
            : const Icon(Icons.login),
        label: const Text('Check in'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('Routes', style: Theme.of(context).textTheme.titleMedium),
          routes.when(
            data: (items) => items.isEmpty
                ? const ListTile(title: Text('No routes'))
                : Column(
                    children: items
                        .map(
                          (r) => ListTile(
                            contentPadding: EdgeInsets.zero,
                            title: Text(r['name']?.toString() ?? ''),
                            subtitle: Text('${r['date'] ?? ''} · ${r['status'] ?? ''}'),
                          ),
                        )
                        .toList(),
                  ),
            loading: () => const LinearProgressIndicator(),
            error: (e, _) => Text('$e'),
          ),
          const SizedBox(height: 16),
          Text('Visits', style: Theme.of(context).textTheme.titleMedium),
          visits.when(
            data: (items) => items.isEmpty
                ? const ListTile(title: Text('No visits'))
                : Column(
                    children: items.map((v) {
                      final open = v['ended_at'] == null;
                      return ListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text('Visit ${v['id']?.toString().substring(0, 8) ?? ''}…'),
                        subtitle: Text('${v['result'] ?? 'in progress'} · customer ${v['customer_id'] ?? ''}'),
                        trailing: open
                            ? TextButton(
                                onPressed: _busy ? null : () => _checkOut(v['id']?.toString() ?? ''),
                                child: const Text('Check out'),
                              )
                            : null,
                      );
                    }).toList(),
                  ),
            loading: () => const LinearProgressIndicator(),
            error: (e, _) => Text('$e'),
          ),
        ],
      ),
    );
  }
}
