import 'package:flutter/material.dart';
import 'package:sfa_app/features/auth/domain/user.dart';

class AdminDestination {
  const AdminDestination({
    required this.path,
    required this.label,
    required this.icon,
    required this.branchIndex,
    this.permission,
  });

  final String path;
  final String label;
  final IconData icon;
  final int branchIndex;
  final String? permission;
}

const adminDestinationCatalog = <AdminDestination>[
  AdminDestination(path: '/dashboard', label: 'Home', icon: Icons.dashboard_outlined, branchIndex: 0),
  AdminDestination(path: '/branches', label: 'Branches', icon: Icons.store_mall_directory_outlined, branchIndex: 1, permission: 'branches:read'),
  AdminDestination(path: '/products', label: 'Products', icon: Icons.inventory_2_outlined, branchIndex: 2, permission: 'catalog:read'),
  AdminDestination(path: '/customers', label: 'Customers', icon: Icons.people_alt_outlined, branchIndex: 3, permission: 'customers:read'),
  AdminDestination(path: '/routes', label: 'Routes', icon: Icons.route_outlined, branchIndex: 4, permission: 'routes:read'),
  AdminDestination(path: '/orders', label: 'Orders', icon: Icons.receipt_long_outlined, branchIndex: 5, permission: 'orders:read'),
  AdminDestination(path: '/returns', label: 'Returns', icon: Icons.assignment_return_outlined, branchIndex: 6, permission: 'returns:read'),
  AdminDestination(path: '/receivables', label: 'Finance', icon: Icons.account_balance_wallet_outlined, branchIndex: 7, permission: 'finance:read'),
  AdminDestination(path: '/documents', label: 'Documents', icon: Icons.folder_outlined, branchIndex: 8, permission: 'documents:read'),
  AdminDestination(path: '/audit', label: 'Audit', icon: Icons.policy_outlined, branchIndex: 9, permission: 'audit:read'),
  AdminDestination(path: '/providers', label: 'Providers', icon: Icons.tune_outlined, branchIndex: 10, permission: 'tenant:read'),
  AdminDestination(path: '/portal-links', label: 'Portal links', icon: Icons.link_outlined, branchIndex: 11, permission: 'portal:write'),
  AdminDestination(path: '/users', label: 'Users', icon: Icons.manage_accounts_outlined, branchIndex: 12, permission: 'users:read'),
  AdminDestination(path: '/roles', label: 'Roles', icon: Icons.security_outlined, branchIndex: 13, permission: 'roles:read'),
  AdminDestination(path: '/branding', label: 'Branding', icon: Icons.palette_outlined, branchIndex: 14, permission: 'tenant:read'),
  AdminDestination(path: '/sync', label: 'Sync', icon: Icons.sync_outlined, branchIndex: 15, permission: 'sync:use'),
  AdminDestination(path: '/notifications', label: 'Alerts', icon: Icons.notifications_outlined, branchIndex: 16, permission: 'notifications:read'),
];

List<AdminDestination> adminDestinationsFor(AuthUser? user) {
  if (user == null) return const [];
  return adminDestinationCatalog.where((d) {
    if (d.permission == null) return true;
    return user.can(d.permission!);
  }).toList();
}
