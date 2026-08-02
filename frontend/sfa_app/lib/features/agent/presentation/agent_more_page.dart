import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

class AgentMorePage extends ConsumerWidget {
  const AgentMorePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(sessionControllerProvider).session?.user;
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Text(user?.fullName ?? 'Agent', style: Theme.of(context).textTheme.headlineSmall),
        Text(user?.email ?? ''),
        const SizedBox(height: 24),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.sync_outlined),
          title: const Text('Sync center'),
          onTap: () => context.push('/field/sync'),
        ),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.notifications_outlined),
          title: const Text('Notifications'),
          onTap: () => context.push('/field/notifications'),
        ),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.assignment_return_outlined),
          title: const Text('Returns'),
          onTap: () => context.push('/field/returns'),
        ),
        const Divider(),
        ListTile(
          contentPadding: EdgeInsets.zero,
          leading: const Icon(Icons.logout_rounded),
          title: const Text('Log out'),
          onTap: () => ref.read(sessionControllerProvider.notifier).logout(),
        ),
      ],
    );
  }
}
