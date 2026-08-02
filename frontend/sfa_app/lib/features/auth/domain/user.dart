class AuthUser {
  const AuthUser({
    required this.id,
    required this.tenantId,
    required this.email,
    required this.fullName,
    required this.locale,
    required this.themePreference,
    required this.roles,
    required this.permissions,
  });

  final String id;
  final String tenantId;
  final String email;
  final String fullName;
  final String locale;
  final String themePreference;
  final List<String> roles;
  final List<String> permissions;

  factory AuthUser.fromJson(Map<String, dynamic> json) {
    return AuthUser(
      id: json['id'] as String,
      tenantId: json['tenant_id'] as String,
      email: json['email'] as String,
      fullName: json['full_name'] as String,
      locale: (json['locale'] as String?) ?? 'ru',
      themePreference: (json['theme_preference'] as String?) ?? 'system',
      roles: (json['roles'] as List<dynamic>? ?? const []).cast<String>(),
      permissions: (json['permissions'] as List<dynamic>? ?? const []).cast<String>(),
    );
  }

  bool get isAdmin =>
      roles.contains('tenant_owner') ||
      roles.contains('tenant_admin') ||
      roles.contains('sales_manager');

  bool get isPortal => roles.contains('customer_portal');

  bool can(String permission) => permissions.contains(permission);
}
