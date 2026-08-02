import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/live/live_channel.dart';
import 'package:sfa_app/core/offline/sync_worker.dart';
import 'package:sfa_app/core/router/app_router.dart';
import 'package:sfa_app/core/theme/brand_theme.dart';
import 'package:sfa_app/features/auth/presentation/auth_controller.dart';

class SfaApp extends ConsumerWidget {
  const SfaApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(sessionControllerProvider);
    final router = ref.watch(appRouterProvider);
    final branding = session.branding ?? Branding.fallback();

    // Keep background sync worker + live channel alive for authenticated sessions.
    if (session.isAuthenticated) {
      ref.watch(syncWorkerProvider);
      ref.watch(liveChannelProvider);
    }

    if (session.loading && session.session == null) {
      return MaterialApp(
        debugShowCheckedModeBanner: false,
        theme: buildBrandTheme(branding, brightness: Brightness.light),
        home: const Scaffold(
          body: Center(child: CircularProgressIndicator()),
        ),
      );
    }

    return MaterialApp.router(
      debugShowCheckedModeBanner: false,
      title: branding.appName,
      theme: buildBrandTheme(branding, brightness: Brightness.light),
      darkTheme: buildBrandTheme(branding, brightness: Brightness.dark),
      themeMode: _themeMode(branding, session.session?.user.themePreference),
      routerConfig: router,
      locale: _locale(session.session?.user.locale),
      supportedLocales: const [
        Locale('ru'),
        Locale('uz'),
        Locale('en'),
      ],
    );
  }

  ThemeMode _themeMode(Branding branding, String? preference) {
    final value = preference ?? branding.themeModeDefault;
    switch (value) {
      case 'dark':
        return ThemeMode.dark;
      case 'light':
        return ThemeMode.light;
      default:
        return ThemeMode.system;
    }
  }

  Locale _locale(String? code) {
    switch (code) {
      case 'uz':
        return const Locale('uz');
      case 'en':
        return const Locale('en');
      default:
        return const Locale('ru');
    }
  }
}
