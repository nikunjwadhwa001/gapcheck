import 'package:flutter/material.dart';

import 'data/models.dart';

/// Material 3 theme generated from one indigo seed colour.
ThemeData buildTheme(Brightness brightness) {
  final scheme = ColorScheme.fromSeed(
    seedColor: Colors.indigo,
    brightness: brightness,
  );
  return ThemeData(
    colorScheme: scheme,
    inputDecorationTheme: const InputDecorationTheme(
      border: OutlineInputBorder(),
      alignLabelWithHint: true,
    ),
    cardTheme: CardThemeData(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(color: scheme.outlineVariant),
      ),
    ),
  );
}

/// Semantic colours for the score and keywords. Kept separate from the
/// indigo brand colour so "good" and "missing" read the same in any theme.
class MatchColors {
  static bool _dark(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark;

  static Color good(BuildContext context) =>
      _dark(context) ? const Color(0xFF6EE7A0) : const Color(0xFF15803D);

  static Color goodSoft(BuildContext context) =>
      _dark(context) ? const Color(0xFF173525) : const Color(0xFFDCF3E3);

  static Color gap(BuildContext context) =>
      _dark(context) ? const Color(0xFFFBBF6A) : const Color(0xFFB45309);

  static Color gapSoft(BuildContext context) =>
      _dark(context) ? const Color(0xFF3D2A12) : const Color(0xFFFDEFD6);

  static Color weak(BuildContext context) =>
      _dark(context) ? const Color(0xFFF87171) : const Color(0xFFB91C1C);

  static Color forBand(BuildContext context, MatchBand band) => switch (band) {
    MatchBand.excellent || MatchBand.strong => good(context),
    MatchBand.partial => gap(context),
    MatchBand.weak => weak(context),
  };
}
