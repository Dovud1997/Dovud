import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

class DashboardPage extends ConsumerWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(sessionControllerProvider);
    final user = state.session?.user;
    final branding = state.branding;
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: Text(branding?.appName ?? 'SFA'),
        actions: [
          IconButton(
            tooltip: 'Log out',
            onPressed: () => ref.read(sessionControllerProvider.notifier).logout(),
            icon: const Icon(Icons.logout_rounded),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          Text(
            'Welcome, ${user?.fullName ?? ''}',
            style: theme.textTheme.headlineMedium,
          ),
          const SizedBox(height: 8),
          Text(
            user?.email ?? '',
            style: theme.textTheme.bodyLarge?.copyWith(
              color: theme.colorScheme.onSurface.withOpacity(0.65),
            ),
          ),
          const SizedBox(height: 24),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final role in user?.roles ?? const <String>[])
                Chip(label: Text(role)),
            ],
          ),
          const SizedBox(height: 28),
          Text('Modules', style: theme.textTheme.titleLarge),
          const SizedBox(height: 12),
          _ModuleTile(
            icon: Icons.store_mall_directory_outlined,
            title: 'Branches',
            subtitle: 'Organization structure',
            onTap: () => context.push('/branches'),
          ),
          _ModuleTile(
            icon: Icons.inventory_2_outlined,
            title: 'Products',
            subtitle: 'Catalog & SKUs',
            onTap: () => context.push('/products'),
          ),
          _ModuleTile(
            icon: Icons.people_alt_outlined,
            title: 'Customers',
            subtitle: 'CRM outlets & contacts',
            onTap: () => context.push('/customers'),
          ),
        ],
      ),
    );
  }
}

class _ModuleTile extends StatelessWidget {
  const _ModuleTile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: ListTile(
        onTap: onTap,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        tileColor: theme.colorScheme.surface,
        leading: Icon(icon, color: theme.colorScheme.primary),
        title: Text(title),
        subtitle: Text(subtitle),
        trailing: const Icon(Icons.chevron_right),
      ),
    );
  }
}
