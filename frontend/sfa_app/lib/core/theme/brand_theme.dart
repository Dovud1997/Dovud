import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class Branding {
  const Branding({
    required this.appName,
    required this.primary,
    required this.secondary,
    required this.accent,
    this.logoUrl,
    this.themeModeDefault = 'light',
    this.version = 1,
  });

  final String appName;
  final Color primary;
  final Color secondary;
  final Color accent;
  final String? logoUrl;
  final String themeModeDefault;
  final int version;

  factory Branding.fallback() => const Branding(
        appName: 'SFA',
        primary: Color(0xFF0F766E),
        secondary: Color(0xFF134E4A),
        accent: Color(0xFFF59E0B),
      );

  factory Branding.fromJson(Map<String, dynamic> json) {
    return Branding(
      appName: (json['app_name'] as String?) ?? 'SFA',
      primary: _parseColor(json['primary_color'] as String?) ?? const Color(0xFF0F766E),
      secondary: _parseColor(json['secondary_color'] as String?) ?? const Color(0xFF134E4A),
      accent: _parseColor(json['accent_color'] as String?) ?? const Color(0xFFF59E0B),
      logoUrl: json['logo_url'] as String?,
      themeModeDefault: (json['theme_mode_default'] as String?) ?? 'light',
      version: (json['branding_version'] as num?)?.toInt() ?? 1,
    );
  }

  static Color? _parseColor(String? hex) {
    if (hex == null || hex.isEmpty) return null;
    var value = hex.replaceAll('#', '');
    if (value.length == 6) value = 'FF$value';
    return Color(int.parse(value, radix: 16));
  }
}

ThemeData buildBrandTheme(Branding branding, {required Brightness brightness}) {
  final base = ColorScheme.fromSeed(
    seedColor: branding.primary,
    brightness: brightness,
    secondary: branding.secondary,
    tertiary: branding.accent,
  );

  final display = GoogleFonts.manropeTextTheme();
  final body = GoogleFonts.plusJakartaSansTextTheme();

  return ThemeData(
    useMaterial3: true,
    colorScheme: base,
    textTheme: body.copyWith(
      displayLarge: display.displayLarge?.copyWith(fontWeight: FontWeight.w700),
      displayMedium: display.displayMedium?.copyWith(fontWeight: FontWeight.w700),
      headlineLarge: display.headlineLarge?.copyWith(fontWeight: FontWeight.w700),
      headlineMedium: display.headlineMedium?.copyWith(fontWeight: FontWeight.w600),
      titleLarge: display.titleLarge?.copyWith(fontWeight: FontWeight.w600),
    ),
    scaffoldBackgroundColor: brightness == Brightness.dark
        ? const Color(0xFF0B1220)
        : const Color(0xFFF3F7F6),
    appBarTheme: AppBarTheme(
      centerTitle: false,
      backgroundColor: Colors.transparent,
      foregroundColor: base.onSurface,
      elevation: 0,
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      border: OutlineInputBorder(borderRadius: BorderRadius.circular(14)),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        minimumSize: const Size.fromHeight(52),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      ),
    ),
  );
}
