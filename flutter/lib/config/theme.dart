import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

// StreamPulse Premium Design System
// Dark-first · Electric Violet · DM Sans
class AppTheme {
  AppTheme._();

  // ── Palette ────────────────────────────────────────────────────────────────

  static const _dark = ColorScheme(
    brightness: Brightness.dark,
    // Primary — electric violet
    primary: Color(0xFFA78BFA),
    onPrimary: Color(0xFF1E0D5E),
    primaryContainer: Color(0xFF3D1F9A),
    onPrimaryContainer: Color(0xFFEDE9FE),
    // Secondary — cherry blossom
    secondary: Color(0xFFF9A8D4),
    onSecondary: Color(0xFF500737),
    secondaryContainer: Color(0xFF881337),
    onSecondaryContainer: Color(0xFFFCE7F3),
    // Tertiary — mint
    tertiary: Color(0xFF6EE7B7),
    onTertiary: Color(0xFF064E3B),
    tertiaryContainer: Color(0xFF065F46),
    onTertiaryContainer: Color(0xFFD1FAE5),
    // Error
    error: Color(0xFFFF6B6B),
    onError: Color(0xFF410002),
    errorContainer: Color(0xFF8C1D18),
    onErrorContainer: Color(0xFFFFDAD6),
    // Surfaces — void black with purple soul
    surface: Color(0xFF0F0E1A),
    onSurface: Color(0xFFE8E6FF),
    surfaceContainerLowest: Color(0xFF07060F),
    surfaceContainerLow: Color(0xFF131222),
    surfaceContainer: Color(0xFF191828),
    surfaceContainerHigh: Color(0xFF1E1D2E),
    surfaceContainerHighest: Color(0xFF252439),
    onSurfaceVariant: Color(0xFFBDBBD9),
    // Structure
    outline: Color(0xFF4A4867),
    outlineVariant: Color(0xFF2D2B45),
    shadow: Color(0xFF000000),
    scrim: Color(0xFF000000),
    inverseSurface: Color(0xFFE8E6FF),
    onInverseSurface: Color(0xFF1A1929),
    inversePrimary: Color(0xFF5B21B6),
  );

  static const _light = ColorScheme(
    brightness: Brightness.light,
    // Primary — deep violet
    primary: Color(0xFF5B21B6),
    onPrimary: Color(0xFFFFFFFF),
    primaryContainer: Color(0xFFEDE9FE),
    onPrimaryContainer: Color(0xFF2E0D7E),
    // Secondary — crimson rose
    secondary: Color(0xFFBE185D),
    onSecondary: Color(0xFFFFFFFF),
    secondaryContainer: Color(0xFFFCE7F3),
    onSecondaryContainer: Color(0xFF831843),
    // Tertiary — emerald
    tertiary: Color(0xFF047857),
    onTertiary: Color(0xFFFFFFFF),
    tertiaryContainer: Color(0xFFD1FAE5),
    onTertiaryContainer: Color(0xFF064E3B),
    // Error
    error: Color(0xFFBA1A1A),
    onError: Color(0xFFFFFFFF),
    errorContainer: Color(0xFFFFDAD6),
    onErrorContainer: Color(0xFF410002),
    // Surfaces — lavender white
    surface: Color(0xFFFAF9FF),
    onSurface: Color(0xFF1A1729),
    surfaceContainerLowest: Color(0xFFFFFFFF),
    surfaceContainerLow: Color(0xFFF5F3FE),
    surfaceContainer: Color(0xFFF0EDFB),
    surfaceContainerHigh: Color(0xFFEBE8F8),
    surfaceContainerHighest: Color(0xFFE5E1F5),
    onSurfaceVariant: Color(0xFF4A4460),
    // Structure
    outline: Color(0xFF7B7392),
    outlineVariant: Color(0xFFCBC4DC),
    shadow: Color(0xFF000000),
    scrim: Color(0xFF000000),
    inverseSurface: Color(0xFF302E45),
    onInverseSurface: Color(0xFFF2EFFE),
    inversePrimary: Color(0xFFC4B5FD),
  );

  // ── Themes ─────────────────────────────────────────────────────────────────

  static ThemeData get dark => _build(_dark);
  static ThemeData get light => _build(_light);

  static ThemeData _build(ColorScheme cs) {
    final base = ThemeData(useMaterial3: true, colorScheme: cs);
    final text = GoogleFonts.dmSansTextTheme(base.textTheme).copyWith(
      displayLarge: GoogleFonts.dmSans(fontSize: 57, fontWeight: FontWeight.w300, letterSpacing: -0.25, color: cs.onSurface),
      displayMedium: GoogleFonts.dmSans(fontSize: 45, fontWeight: FontWeight.w300, color: cs.onSurface),
      displaySmall: GoogleFonts.dmSans(fontSize: 36, fontWeight: FontWeight.w400, color: cs.onSurface),
      headlineLarge: GoogleFonts.dmSans(fontSize: 32, fontWeight: FontWeight.w600, letterSpacing: -0.5, color: cs.onSurface),
      headlineMedium: GoogleFonts.dmSans(fontSize: 28, fontWeight: FontWeight.w600, letterSpacing: -0.25, color: cs.onSurface),
      headlineSmall: GoogleFonts.dmSans(fontSize: 24, fontWeight: FontWeight.w600, color: cs.onSurface),
      titleLarge: GoogleFonts.dmSans(fontSize: 20, fontWeight: FontWeight.w600, color: cs.onSurface),
      titleMedium: GoogleFonts.dmSans(fontSize: 16, fontWeight: FontWeight.w500, letterSpacing: 0.1, color: cs.onSurface),
      titleSmall: GoogleFonts.dmSans(fontSize: 14, fontWeight: FontWeight.w500, letterSpacing: 0.1, color: cs.onSurface),
      bodyLarge: GoogleFonts.dmSans(fontSize: 16, fontWeight: FontWeight.w400, color: cs.onSurface),
      bodyMedium: GoogleFonts.dmSans(fontSize: 14, fontWeight: FontWeight.w400, color: cs.onSurface),
      bodySmall: GoogleFonts.dmSans(fontSize: 12, fontWeight: FontWeight.w400, color: cs.onSurfaceVariant),
      labelLarge: GoogleFonts.dmSans(fontSize: 14, fontWeight: FontWeight.w600, letterSpacing: 0.1, color: cs.onSurface),
      labelMedium: GoogleFonts.dmSans(fontSize: 12, fontWeight: FontWeight.w500, letterSpacing: 0.5, color: cs.onSurface),
      labelSmall: GoogleFonts.dmSans(fontSize: 11, fontWeight: FontWeight.w500, letterSpacing: 0.5, color: cs.onSurface),
    );

    return base.copyWith(
      brightness: cs.brightness,
      scaffoldBackgroundColor: cs.surfaceContainerLowest,
      textTheme: text,

      appBarTheme: AppBarTheme(
        backgroundColor: cs.surfaceContainerLowest,
        foregroundColor: cs.onSurface,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
        titleTextStyle: GoogleFonts.dmSans(
          fontSize: 20,
          fontWeight: FontWeight.w700,
          color: cs.onSurface,
        ),
      ),

      navigationBarTheme: NavigationBarThemeData(
        height: 72,
        backgroundColor: cs.surface,
        surfaceTintColor: Colors.transparent,
        shadowColor: Colors.transparent,
        indicatorColor: cs.primaryContainer,
        labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
        labelTextStyle: WidgetStateProperty.resolveWith((states) {
          final selected = states.contains(WidgetState.selected);
          return GoogleFonts.dmSans(
            fontSize: 11,
            fontWeight: selected ? FontWeight.w600 : FontWeight.w400,
            color: selected ? cs.primary : cs.onSurfaceVariant,
          );
        }),
        iconTheme: WidgetStateProperty.resolveWith((states) {
          final selected = states.contains(WidgetState.selected);
          return IconThemeData(
            color: selected ? cs.primary : cs.onSurfaceVariant,
            size: 24,
          );
        }),
      ),

      cardTheme: CardThemeData(
        color: cs.surfaceContainer,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(color: cs.outlineVariant, width: 0.5),
        ),
        margin: EdgeInsets.zero,
      ),

      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: cs.primary,
          foregroundColor: cs.onPrimary,
          minimumSize: const Size(0, 52),
          padding: const EdgeInsets.symmetric(horizontal: 24),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          textStyle: GoogleFonts.dmSans(fontSize: 15, fontWeight: FontWeight.w600),
        ),
      ),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: cs.surfaceContainerHigh,
          foregroundColor: cs.onSurface,
          minimumSize: const Size(0, 52),
          padding: const EdgeInsets.symmetric(horizontal: 24),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          elevation: 0,
          textStyle: GoogleFonts.dmSans(fontSize: 15, fontWeight: FontWeight.w500),
        ),
      ),

      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: cs.primary,
          side: BorderSide(color: cs.outline),
          minimumSize: const Size(0, 52),
          padding: const EdgeInsets.symmetric(horizontal: 24),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
          textStyle: GoogleFonts.dmSans(fontSize: 15, fontWeight: FontWeight.w500),
        ),
      ),

      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: cs.primary,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
          textStyle: GoogleFonts.dmSans(fontSize: 14, fontWeight: FontWeight.w600),
        ),
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: cs.surfaceContainerHigh,
        contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide.none,
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: cs.outlineVariant, width: 0.5),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: cs.primary, width: 1.5),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(14),
          borderSide: BorderSide(color: cs.error, width: 1),
        ),
        labelStyle: GoogleFonts.dmSans(color: cs.onSurfaceVariant, fontSize: 14),
        hintStyle: GoogleFonts.dmSans(color: cs.onSurfaceVariant, fontSize: 14),
        floatingLabelStyle: GoogleFonts.dmSans(color: cs.primary, fontSize: 12, fontWeight: FontWeight.w600),
      ),

      chipTheme: ChipThemeData(
        backgroundColor: cs.surfaceContainerHigh,
        labelStyle: GoogleFonts.dmSans(fontSize: 12, fontWeight: FontWeight.w500),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        side: BorderSide(color: cs.outlineVariant, width: 0.5),
      ),

      dividerTheme: DividerThemeData(
        color: cs.outlineVariant,
        thickness: 0.5,
        space: 0,
      ),

      sliderTheme: SliderThemeData(
        activeTrackColor: cs.primary,
        inactiveTrackColor: cs.surfaceContainerHighest,
        thumbColor: cs.primary,
        overlayColor: cs.primary.withValues(alpha: 0.12),
        trackHeight: 3,
        thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6),
      ),

      listTileTheme: ListTileThemeData(
        contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 4),
        iconColor: cs.onSurfaceVariant,
        titleTextStyle: GoogleFonts.dmSans(fontSize: 14, fontWeight: FontWeight.w500, color: cs.onSurface),
        subtitleTextStyle: GoogleFonts.dmSans(fontSize: 12, color: cs.onSurfaceVariant),
      ),

      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: cs.primary,
        foregroundColor: cs.onPrimary,
        elevation: 4,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
      ),
    );
  }
}
