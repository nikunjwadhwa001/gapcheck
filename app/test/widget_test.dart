import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gapcheck/data/models.dart';
import 'package:gapcheck/main.dart';
import 'package:gapcheck/screens/results_screen.dart';
import 'package:gapcheck/state/analysis_controller.dart';

/// A controller that starts in a fixed state, so screens can be tested
/// without any network calls.
class FakeController extends AnalysisController {
  FakeController(this.initial);
  final AsyncValue<AnalysisResult>? initial;

  @override
  AsyncValue<AnalysisResult>? build() => initial;
}

Widget resultsWith(AsyncValue<AnalysisResult>? state) => ProviderScope(
  overrides: [analysisProvider.overrideWith(() => FakeController(state))],
  child: const MaterialApp(home: ResultsScreen()),
);

const result = AnalysisResult(
  score: 62,
  summary: 'Solid Go background, missing DevOps tools.',
  matchedKeywords: ['Go', 'PostgreSQL'],
  missingKeywords: ['Docker', 'Kubernetes', 'gRPC'],
  suggestions: [
    Suggestion(
      section: 'Skills',
      issue: 'No containers listed',
      fix: 'Add Docker if you have used it.',
    ),
  ],
);

void main() {
  group('input screen', () {
    Future<void> pumpApp(WidgetTester tester) async {
      // A desktop-sized window, so the whole form is on screen.
      tester.view.physicalSize = const Size(1200, 1600);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(const ProviderScope(child: GapCheckApp()));
      await tester.pumpAndSettle();
    }

    FilledButton analyzeButton(WidgetTester tester) => tester.widget(
      find.ancestor(
        of: find.text('Analyze match'),
        matching: find.byWidgetPredicate((w) => w is FilledButton),
      ),
    );

    testWidgets('starts on the input screen with Analyze disabled', (
      tester,
    ) async {
      await pumpApp(tester);

      expect(
        find.text('How well does your resume fit the job?'),
        findsOneWidget,
      );
      expect(analyzeButton(tester).onPressed, isNull);
    });

    testWidgets('Analyze enables only when both texts are long enough', (
      tester,
    ) async {
      await pumpApp(tester);
      final fields = find.byType(TextField);

      await tester.enterText(fields.at(0), 'x' * minTextLength);
      await tester.pump();
      expect(analyzeButton(tester).onPressed, isNull);
      expect(find.textContaining('more)'), findsNothing);

      await tester.enterText(fields.at(1), 'too short');
      await tester.pump();
      expect(analyzeButton(tester).onPressed, isNull);
      expect(find.textContaining('41 more'), findsOneWidget);

      await tester.enterText(fields.at(1), 'y' * minTextLength);
      await tester.pump();
      expect(analyzeButton(tester).onPressed, isNotNull);
    });

    testWidgets('Try an example fills both boxes', (tester) async {
      await pumpApp(tester);

      await tester.tap(find.text('Try an example'));
      await tester.pump();

      expect(find.textContaining('Priya Sharma'), findsOneWidget);
      expect(analyzeButton(tester).onPressed, isNotNull);
    });
  });

  group('results screen', () {
    testWidgets('shows score, band, keywords and suggestions', (tester) async {
      await tester.pumpWidget(resultsWith(const AsyncData(result)));
      await tester.pumpAndSettle(); // let the score ring finish animating

      expect(find.text('62'), findsOneWidget);
      expect(find.text('Partial match'), findsOneWidget);
      expect(find.text('Missing from your resume (3)'), findsOneWidget);
      expect(find.text('Docker'), findsOneWidget);
      expect(find.text('Already matching (2)'), findsOneWidget);
      expect(find.text('Add Docker if you have used it.'), findsOneWidget);
    });

    testWidgets('shows a loading message while waiting', (tester) async {
      await tester.pumpWidget(resultsWith(const AsyncLoading()));
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Comparing your resume with the job…'), findsOneWidget);
    });

    testWidgets('rate limit error shows its title, message and retry', (
      tester,
    ) async {
      const error = ApiException(
        code: 'rate_limited',
        message: 'Too many analyses right now.',
      );
      await tester.pumpWidget(
        resultsWith(const AsyncError(error, StackTrace.empty)),
      );
      await tester.pump();

      expect(find.text('Too many requests'), findsOneWidget);
      expect(find.text('Too many analyses right now.'), findsOneWidget);
      expect(find.text('Try again'), findsOneWidget);
    });

    testWidgets('input errors offer editing, not retrying', (tester) async {
      const error = ApiException(
        code: 'invalid_input',
        message: 'jd_text is required',
        field: 'jd_text',
      );
      await tester.pumpWidget(
        resultsWith(const AsyncError(error, StackTrace.empty)),
      );
      await tester.pump();

      expect(find.text('Check your input'), findsOneWidget);
      expect(find.text('Try again'), findsNothing);
      expect(find.text('Edit inputs'), findsOneWidget);
    });

    testWidgets('opening results directly shows a way back', (tester) async {
      await tester.pumpWidget(resultsWith(null));
      await tester.pump();

      expect(find.text('No analysis yet.'), findsOneWidget);
    });
  });
}
