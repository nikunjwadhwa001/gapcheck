import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../data/models.dart';
import '../data/sample.dart';
import '../state/analysis_controller.dart';

/// Screen 1: paste a resume and a job description, then analyze.
class InputScreen extends ConsumerStatefulWidget {
  const InputScreen({super.key});

  @override
  ConsumerState<InputScreen> createState() => _InputScreenState();
}

class _InputScreenState extends ConsumerState<InputScreen> {
  final _resume = TextEditingController();
  final _jd = TextEditingController();

  @override
  void initState() {
    super.initState();
    // Start waking the free-tier server now; pasting takes longer than its
    // ~1 minute cold start, so the first analysis is usually fast.
    unawaited(ref.read(apiProvider).wakeUp());
  }

  @override
  void dispose() {
    _resume.dispose();
    _jd.dispose();
    super.dispose();
  }

  bool get _canAnalyze =>
      _isValid(_resume.text) &&
      _isValid(_jd.text) &&
      ref.read(analysisProvider) is! AsyncLoading;

  static bool _isValid(String text) {
    final n = text.trim().length;
    return n >= minTextLength && n <= maxTextLength;
  }

  void _analyze() {
    ref
        .read(analysisProvider.notifier)
        .analyze(resume: _resume.text.trim(), jd: _jd.text.trim());
    context.push('/results');
  }

  void _loadExample() {
    _resume.text = sampleResume;
    _jd.text = sampleJobDescription;
  }

  void _clear() {
    _resume.clear();
    _jd.clear();
  }

  @override
  Widget build(BuildContext context) {
    // Rebuild when the analysis starts or finishes, so the button state is right.
    ref.watch(analysisProvider);
    final theme = Theme.of(context);

    final resumeBox = _TextBox(
      controller: _resume,
      label: 'Your resume',
      hint: 'Paste the text of your resume…',
    );
    final jdBox = _TextBox(
      controller: _jd,
      label: 'Job description',
      hint: 'Paste the job posting…',
    );

    return Scaffold(
      appBar: AppBar(title: const Text('GapCheck')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1100),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Text(
                  'How well does your resume fit the job?',
                  style: theme.textTheme.headlineSmall,
                ),
                const SizedBox(height: 4),
                Text(
                  'Paste both texts. GapCheck scores the match, lists the keywords '
                  "you're missing, and suggests specific edits.",
                  style: theme.textTheme.bodyMedium?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 20),
                LayoutBuilder(
                  builder: (context, constraints) {
                    // Side by side on wide screens, stacked on phones.
                    if (constraints.maxWidth >= 760) {
                      return Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(child: resumeBox),
                          const SizedBox(width: 16),
                          Expanded(child: jdBox),
                        ],
                      );
                    }
                    return Column(
                      children: [resumeBox, const SizedBox(height: 16), jdBox],
                    );
                  },
                ),
                const SizedBox(height: 20),
                Wrap(
                  spacing: 12,
                  runSpacing: 8,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    ListenableBuilder(
                      listenable: Listenable.merge([_resume, _jd]),
                      builder: (context, _) => FilledButton.icon(
                        onPressed: _canAnalyze ? _analyze : null,
                        icon: const Icon(Icons.insights),
                        label: const Text('Analyze match'),
                      ),
                    ),
                    TextButton(
                      onPressed: _loadExample,
                      child: const Text('Try an example'),
                    ),
                    TextButton(onPressed: _clear, child: const Text('Clear')),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// A multi-line text box with a live character count and length hint.
class _TextBox extends StatelessWidget {
  const _TextBox({
    required this.controller,
    required this.label,
    required this.hint,
  });

  final TextEditingController controller;
  final String label;
  final String hint;

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: controller,
      builder: (context, _) {
        final n = controller.text.trim().length;
        final String? error = n > maxTextLength
            ? 'Too long: keep it under ${_fmt(maxTextLength)} characters'
            : null;
        final String? helper = n == 0
            ? null
            : n < minTextLength
            ? 'At least $minTextLength characters (${minTextLength - n} more)'
            : null;

        return TextField(
          controller: controller,
          minLines: 12,
          maxLines: 18,
          keyboardType: TextInputType.multiline,
          decoration: InputDecoration(
            labelText: label,
            hintText: hint,
            helperText: helper,
            errorText: error,
            counterText: '${_fmt(n)} / ${_fmt(maxTextLength)}',
          ),
        );
      },
    );
  }
}

/// 12345 → "12,345". Small enough not to need the intl package.
String _fmt(int n) =>
    n.toString().replaceAllMapped(RegExp(r'\B(?=(\d{3})+(?!\d))'), (_) => ',');
