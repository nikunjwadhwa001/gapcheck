import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../data/models.dart';
import '../state/analysis_controller.dart';
import '../theme.dart';
import '../widgets/error_view.dart';
import '../widgets/keyword_section.dart';
import '../widgets/score_ring.dart';
import '../widgets/suggestion_card.dart';

/// Screen 2: shows loading, an error, or the analysis.
class ResultsScreen extends ConsumerWidget {
  const ResultsScreen({super.key});

  void _back(BuildContext context) =>
      context.canPop() ? context.pop() : context.go('/');

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final analysis = ref.watch(analysisProvider);

    final Widget body = switch (analysis) {
      null => _NoAnalysis(onStart: () => context.go('/')),
      AsyncLoading() => const _Loading(),
      AsyncError(:final error) => ErrorView(
        error: error,
        onRetry: () => ref.read(analysisProvider.notifier).retry(),
        onEdit: () => _back(context),
      ),
      AsyncData(:final value) => _ResultView(
        result: value,
        onAgain: () => _back(context),
      ),
    };

    return Scaffold(
      appBar: AppBar(
        title: const Text('Your match'),
        leading: BackButton(onPressed: () => _back(context)),
      ),
      body: SafeArea(child: body),
    );
  }
}

class _Loading extends StatelessWidget {
  const _Loading();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const CircularProgressIndicator(),
            const SizedBox(height: 20),
            Text(
              'Comparing your resume with the job…',
              style: theme.textTheme.titleMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 4),
            Text(
              'This usually takes a few seconds.',
              style: theme.textTheme.bodySmall,
            ),
          ],
        ),
      ),
    );
  }
}

class _NoAnalysis extends StatelessWidget {
  const _NoAnalysis({required this.onStart});

  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text('No analysis yet.'),
          const SizedBox(height: 12),
          FilledButton(onPressed: onStart, child: const Text('Paste a resume')),
        ],
      ),
    );
  }
}

class _ResultView extends StatelessWidget {
  const _ResultView({required this.result, required this.onAgain});

  final AnalysisResult result;
  final VoidCallback onAgain;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final bandColor = MatchColors.forBand(context, result.band);

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 860),
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Score + summary
            Card(
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: Wrap(
                  spacing: 24,
                  runSpacing: 16,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    ScoreRing(score: result.score, color: bandColor),
                    ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 520),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            result.band.label,
                            style: theme.textTheme.titleLarge?.copyWith(
                              color: bandColor,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            result.summary,
                            style: theme.textTheme.bodyLarge,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
            KeywordSection(
              title: 'Missing from your resume',
              keywords: result.missingKeywords,
              kind: KeywordKind.missing,
              emptyText: 'Nothing important is missing.',
            ),
            const SizedBox(height: 20),
            KeywordSection(
              title: 'Already matching',
              keywords: result.matchedKeywords,
              kind: KeywordKind.matched,
              emptyText: 'No clear matches found yet.',
            ),
            const SizedBox(height: 28),
            Text('Suggested edits', style: theme.textTheme.titleMedium),
            const SizedBox(height: 8),
            if (result.suggestions.isEmpty)
              const Text('No edits suggested.')
            else
              for (final s in result.suggestions) SuggestionCard(suggestion: s),
            const SizedBox(height: 24),
            Align(
              alignment: Alignment.centerLeft,
              child: OutlinedButton.icon(
                onPressed: onAgain,
                icon: const Icon(Icons.edit_note),
                label: const Text('Edit and analyze again'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
