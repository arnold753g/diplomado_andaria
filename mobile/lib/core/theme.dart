import 'package:flutter/material.dart';

const lime = Color(0xFFBAFE06);
const ink = Color(0xFF080C0F);
const muted = Color(0xFF62686B);
const canvas = Color(0xFFF7F8F5);
const line = Color(0xFFE4E7E1);
final andariaTheme = ThemeData(
  useMaterial3: true,
  fontFamily: 'Outfit',
  scaffoldBackgroundColor: Colors.white,
  colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF567400))
      .copyWith(
        primary: const Color(0xFF466000),
        onPrimary: Colors.white,
        secondary: lime,
        onSecondary: ink,
        surface: Colors.white,
        onSurface: ink,
      ),
  appBarTheme: const AppBarTheme(
    backgroundColor: Colors.white,
    foregroundColor: ink,
    centerTitle: false,
    scrolledUnderElevation: 0,
  ),
  dividerTheme: const DividerThemeData(color: line, space: 32),
  chipTheme: ChipThemeData(
    backgroundColor: canvas,
    selectedColor: lime,
    side: const BorderSide(color: line),
    labelStyle: const TextStyle(fontFamily: 'Outfit', color: ink),
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 10),
  ),
  inputDecorationTheme: InputDecorationTheme(
    filled: true,
    fillColor: canvas,
    contentPadding: const EdgeInsets.all(16),
    border: OutlineInputBorder(
      borderRadius: BorderRadius.circular(14),
      borderSide: const BorderSide(color: line),
    ),
    enabledBorder: OutlineInputBorder(
      borderRadius: BorderRadius.circular(14),
      borderSide: const BorderSide(color: line),
    ),
  ),
  filledButtonTheme: FilledButtonThemeData(
    style: FilledButton.styleFrom(
      backgroundColor: lime,
      foregroundColor: ink,
      minimumSize: const Size(48, 52),
      textStyle: const TextStyle(
        fontFamily: 'Outfit',
        fontSize: 16,
        fontWeight: FontWeight.w600,
      ),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
    ),
  ),
  outlinedButtonTheme: OutlinedButtonThemeData(
    style: OutlinedButton.styleFrom(
      foregroundColor: ink,
      minimumSize: const Size(48, 48),
      side: const BorderSide(color: line),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
    ),
  ),
  navigationBarTheme: NavigationBarThemeData(
    backgroundColor: Colors.white,
    indicatorColor: lime,
    height: 76,
    labelTextStyle: WidgetStateProperty.all(
      const TextStyle(
        fontFamily: 'Outfit',
        fontSize: 12,
        fontWeight: FontWeight.w500,
      ),
    ),
  ),
  textTheme: const TextTheme(
    headlineLarge: TextStyle(
      fontSize: 34,
      fontWeight: FontWeight.w700,
      height: 1.1,
      letterSpacing: -1.2,
    ),
    headlineMedium: TextStyle(
      fontSize: 28,
      fontWeight: FontWeight.w700,
      height: 1.15,
      letterSpacing: -.6,
    ),
    titleLarge: TextStyle(fontSize: 22, fontWeight: FontWeight.w600),
    titleMedium: TextStyle(fontSize: 17, fontWeight: FontWeight.w600),
    bodyLarge: TextStyle(fontSize: 16, height: 1.5),
    bodyMedium: TextStyle(fontSize: 14, height: 1.5),
    bodySmall: TextStyle(fontSize: 12, height: 1.4, color: muted),
  ).apply(bodyColor: ink, displayColor: ink),
);
