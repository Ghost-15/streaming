import 'package:flutter/material.dart';

/// Material 3 Theme configuration for StreamPulse
class AppTheme {
  // Prevent instantiation
  AppTheme._();

  // Color palette (seed color: Purple for audio/media)
  static const Color _seedColor = Color(0xFF7C3AED); // Violet
  static const Color _darkSeedColor = Color(0xFFC4B5FD); // Light Violet

  /// Light theme
  static ThemeData get light {
    final colorScheme = ColorScheme.fromSeed(
      seedColor: _seedColor,
      brightness: Brightness.light,
      // Customize dynamic colors if needed
      primaryContainer: const Color(0xFFEDE7FF),
      secondaryContainer: const Color(0xFFFFD8F4),
      tertiaryContainer: const Color(0xFFFFF8E0),
      surfaceDim: const Color(0xFFE8E5F0),
      surfaceBright: const Color(0xFFFAF8FF),
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      brightness: Brightness.light,
      scaffoldBackgroundColor: colorScheme.surface,

      // AppBar theme
      appBarTheme: AppBarTheme(
        backgroundColor: colorScheme.surface,
        foregroundColor: colorScheme.onSurface,
        elevation: 0,
        centerTitle: false,
        titleTextStyle: _headlineSmall(colorScheme.onSurface),
      ),

      // Bottom App Bar theme
      bottomAppBarTheme: BottomAppBarThemeData(
        color: colorScheme.surface,
        elevation: 1,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      ),

      // Navigation Bar theme
      navigationBarTheme: NavigationBarThemeData(
        height: 80,
        labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
        backgroundColor: colorScheme.surface,
        elevation: 1,
        indicatorColor: colorScheme.primaryContainer,
        labelTextStyle: MaterialStateProperty.all(
          _labelSmall(colorScheme.onSurface),
        ),
      ),

      // FAB theme
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: colorScheme.primary,
        foregroundColor: colorScheme.onPrimary,
        elevation: 4,
        sizeConstraints: const BoxConstraints(minWidth: 56, minHeight: 56),
      ),

      // Button themes
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: colorScheme.primary,
          foregroundColor: colorScheme.onPrimary,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          elevation: 2,
          textStyle: _labelLarge(colorScheme.onPrimary),
        ),
      ),

      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: colorScheme.primary,
          side: BorderSide(color: colorScheme.primary),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: _labelLarge(colorScheme.primary),
        ),
      ),

      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: colorScheme.primary,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle: _labelLarge(colorScheme.primary),
        ),
      ),

      // Card theme
      cardTheme: CardThemeData(
        color: colorScheme.surface,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(8),
      ),

      // Input decoration theme
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colorScheme.surfaceContainerHighest,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.outline),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.outlineVariant),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.primary, width: 2),
        ),
        labelStyle: _bodyMedium(colorScheme.onSurfaceVariant),
        hintStyle: _bodyMedium(colorScheme.onSurfaceVariant),
      ),

      // Text themes
      textTheme: TextTheme(
        displayLarge: _displayLarge(colorScheme.onSurface),
        displayMedium: _displayMedium(colorScheme.onSurface),
        displaySmall: _displaySmall(colorScheme.onSurface),
        headlineLarge: _headlineLarge(colorScheme.onSurface),
        headlineMedium: _headlineMedium(colorScheme.onSurface),
        headlineSmall: _headlineSmall(colorScheme.onSurface),
        titleLarge: _titleLarge(colorScheme.onSurface),
        titleMedium: _titleMedium(colorScheme.onSurface),
        titleSmall: _titleSmall(colorScheme.onSurface),
        bodyLarge: _bodyLarge(colorScheme.onSurface),
        bodyMedium: _bodyMedium(colorScheme.onSurface),
        bodySmall: _bodySmall(colorScheme.onSurface),
        labelLarge: _labelLarge(colorScheme.onSurface),
        labelMedium: _labelMedium(colorScheme.onSurface),
        labelSmall: _labelSmall(colorScheme.onSurface),
      ),

      // Divider theme
      dividerTheme: DividerThemeData(
        color: colorScheme.outlineVariant,
        thickness: 1,
        space: 16,
      ),

      // Slider theme
      sliderTheme: SliderThemeData(
        activeTrackColor: colorScheme.primary,
        inactiveTrackColor: colorScheme.surfaceContainerHighest,
        thumbColor: colorScheme.primary,
        overlayColor: colorScheme.primary.withOpacity(0.12),
        trackHeight: 4,
      ),
    );
  }

  /// Dark theme
  static ThemeData get dark {
    final colorScheme = ColorScheme.fromSeed(
      seedColor: _darkSeedColor,
      brightness: Brightness.dark,
      primaryContainer: const Color(0xFF6200EE),
      secondaryContainer: const Color(0xFF6A1B9A),
      tertiaryContainer: const Color(0xFFB39DDB),
      surfaceDim: const Color(0xFF12121C),
      surfaceBright: const Color(0xFF3A3A47),
    );

    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      brightness: Brightness.dark,
      scaffoldBackgroundColor: colorScheme.surface,

      // AppBar theme
      appBarTheme: AppBarTheme(
        backgroundColor: colorScheme.surface,
        foregroundColor: colorScheme.onSurface,
        elevation: 0,
        centerTitle: false,
        titleTextStyle: _headlineSmall(colorScheme.onSurface),
      ),

      // Bottom App Bar theme
      bottomAppBarTheme: BottomAppBarThemeData(
        color: colorScheme.surface,
        elevation: 1,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      ),

      // Navigation Bar theme
      navigationBarTheme: NavigationBarThemeData(
        height: 80,
        labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
        backgroundColor: colorScheme.surface,
        elevation: 1,
        indicatorColor: colorScheme.primaryContainer,
        labelTextStyle: MaterialStateProperty.all(
          _labelSmall(colorScheme.onSurface),
        ),
      ),

      // FAB theme
      floatingActionButtonTheme: FloatingActionButtonThemeData(
        backgroundColor: colorScheme.primary,
        foregroundColor: colorScheme.onPrimary,
        elevation: 4,
        sizeConstraints: const BoxConstraints(minWidth: 56, minHeight: 56),
      ),

      // Button themes
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: colorScheme.primary,
          foregroundColor: colorScheme.onPrimary,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          elevation: 2,
          textStyle: _labelLarge(colorScheme.onPrimary),
        ),
      ),

      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: colorScheme.primary,
          side: BorderSide(color: colorScheme.primary),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: _labelLarge(colorScheme.primary),
        ),
      ),

      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: colorScheme.primary,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle: _labelLarge(colorScheme.primary),
        ),
      ),

      // Card theme
      cardTheme: CardThemeData(
        color: colorScheme.surface,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(8),
      ),

      // Input decoration theme
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colorScheme.surfaceContainerHighest,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.outline),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.outlineVariant),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: colorScheme.primary, width: 2),
        ),
        labelStyle: _bodyMedium(colorScheme.onSurfaceVariant),
        hintStyle: _bodyMedium(colorScheme.onSurfaceVariant),
      ),

      // Text themes
      textTheme: TextTheme(
        displayLarge: _displayLarge(colorScheme.onSurface),
        displayMedium: _displayMedium(colorScheme.onSurface),
        displaySmall: _displaySmall(colorScheme.onSurface),
        headlineLarge: _headlineLarge(colorScheme.onSurface),
        headlineMedium: _headlineMedium(colorScheme.onSurface),
        headlineSmall: _headlineSmall(colorScheme.onSurface),
        titleLarge: _titleLarge(colorScheme.onSurface),
        titleMedium: _titleMedium(colorScheme.onSurface),
        titleSmall: _titleSmall(colorScheme.onSurface),
        bodyLarge: _bodyLarge(colorScheme.onSurface),
        bodyMedium: _bodyMedium(colorScheme.onSurface),
        bodySmall: _bodySmall(colorScheme.onSurface),
        labelLarge: _labelLarge(colorScheme.onSurface),
        labelMedium: _labelMedium(colorScheme.onSurface),
        labelSmall: _labelSmall(colorScheme.onSurface),
      ),

      // Divider theme
      dividerTheme: DividerThemeData(
        color: colorScheme.outlineVariant,
        thickness: 1,
        space: 16,
      ),

      // Slider theme
      sliderTheme: SliderThemeData(
        activeTrackColor: colorScheme.primary,
        inactiveTrackColor: colorScheme.surfaceContainerHighest,
        thumbColor: colorScheme.primary,
        overlayColor: colorScheme.primary.withOpacity(0.12),
        trackHeight: 4,
      ),
    );
  }

  // Typography helpers (Material 3)

  static TextStyle _displayLarge(Color color) => TextStyle(
        fontSize: 57,
        fontWeight: FontWeight.w400,
        height: 1.12,
        letterSpacing: -0.25,
        color: color,
      );

  static TextStyle _displayMedium(Color color) => TextStyle(
        fontSize: 45,
        fontWeight: FontWeight.w400,
        height: 1.16,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _displaySmall(Color color) => TextStyle(
        fontSize: 36,
        fontWeight: FontWeight.w400,
        height: 1.22,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _headlineLarge(Color color) => TextStyle(
        fontSize: 32,
        fontWeight: FontWeight.w400,
        height: 1.25,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _headlineMedium(Color color) => TextStyle(
        fontSize: 28,
        fontWeight: FontWeight.w500,
        height: 1.29,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _headlineSmall(Color color) => TextStyle(
        fontSize: 24,
        fontWeight: FontWeight.w500,
        height: 1.33,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _titleLarge(Color color) => TextStyle(
        fontSize: 22,
        fontWeight: FontWeight.w500,
        height: 1.27,
        letterSpacing: 0,
        color: color,
      );

  static TextStyle _titleMedium(Color color) => TextStyle(
        fontSize: 16,
        fontWeight: FontWeight.w500,
        height: 1.5,
        letterSpacing: 0.15,
        color: color,
      );

  static TextStyle _titleSmall(Color color) => TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        height: 1.43,
        letterSpacing: 0.1,
        color: color,
      );

  static TextStyle _bodyLarge(Color color) => TextStyle(
        fontSize: 16,
        fontWeight: FontWeight.w400,
        height: 1.5,
        letterSpacing: 0.15,
        color: color,
      );

  static TextStyle _bodyMedium(Color color) => TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w400,
        height: 1.43,
        letterSpacing: 0.25,
        color: color,
      );

  static TextStyle _bodySmall(Color color) => TextStyle(
        fontSize: 12,
        fontWeight: FontWeight.w400,
        height: 1.33,
        letterSpacing: 0.4,
        color: color,
      );

  static TextStyle _labelLarge(Color color) => TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w500,
        height: 1.43,
        letterSpacing: 0.1,
        color: color,
      );

  static TextStyle _labelMedium(Color color) => TextStyle(
        fontSize: 12,
        fontWeight: FontWeight.w500,
        height: 1.33,
        letterSpacing: 0.5,
        color: color,
      );

  static TextStyle _labelSmall(Color color) => TextStyle(
        fontSize: 11,
        fontWeight: FontWeight.w500,
        height: 1.45,
        letterSpacing: 0.5,
        color: color,
      );
}
