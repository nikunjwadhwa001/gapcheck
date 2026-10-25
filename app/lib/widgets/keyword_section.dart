import 'package:flutter/material.dart';

import '../theme.dart';

enum KeywordKind { matched, missing }

/// A titled group of keyword chips: green for matched, amber for missing.
class KeywordSection extends StatelessWidget {
  const KeywordSection({
    super.key,
    required this.title,
    required this.keywords,
    required this.kind,
    required this.emptyText,
  });

  final String title;
  final List<String> keywords;
  final KeywordKind kind;
  final String emptyText;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final missing = kind == KeywordKind.missing;
    final fg = missing ? MatchColors.gap(context) : MatchColors.good(context);
    final bg = missing
        ? MatchColors.gapSoft(context)
        : MatchColors.goodSoft(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('$title (${keywords.length})', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        if (keywords.isEmpty)
          Text(emptyText, style: theme.textTheme.bodyMedium)
        else
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final k in keywords)
                Chip(
                  avatar: Icon(
                    missing ? Icons.add_circle_outline : Icons.check_circle,
                    size: 18,
                    color: fg,
                  ),
                  label: Text(k),
                  labelStyle: TextStyle(color: fg, fontWeight: FontWeight.w600),
                  backgroundColor: bg,
                  side: BorderSide.none,
                ),
            ],
          ),
      ],
    );
  }
}
