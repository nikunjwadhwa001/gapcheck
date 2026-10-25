import 'package:flutter_test/flutter_test.dart';
import 'package:gapcheck/data/models.dart';

void main() {
  test('AnalysisResult.fromJson reads every field', () {
    final r = AnalysisResult.fromJson({
      'score': 72,
      'summary': 'Good fit.',
      'matched_keywords': ['Go', 'PostgreSQL'],
      'missing_keywords': ['Docker'],
      'suggestions': [
        {'section': 'Skills', 'issue': 'No containers', 'fix': 'Add Docker.'},
      ],
    });

    expect(r.score, 72);
    expect(r.summary, 'Good fit.');
    expect(r.matchedKeywords, ['Go', 'PostgreSQL']);
    expect(r.missingKeywords, ['Docker']);
    expect(r.suggestions.single.fix, 'Add Docker.');
    expect(r.band, MatchBand.strong);
  });

  test('AnalysisResult.fromJson tolerates missing or null lists', () {
    final r = AnalysisResult.fromJson({'score': 10, 'matched_keywords': null});

    expect(r.matchedKeywords, isEmpty);
    expect(r.missingKeywords, isEmpty);
    expect(r.suggestions, isEmpty);
    expect(r.summary, '');
  });

  test('MatchBand uses the same thresholds as the backend prompt', () {
    expect(MatchBand.fromScore(100), MatchBand.excellent);
    expect(MatchBand.fromScore(90), MatchBand.excellent);
    expect(MatchBand.fromScore(89), MatchBand.strong);
    expect(MatchBand.fromScore(70), MatchBand.strong);
    expect(MatchBand.fromScore(69), MatchBand.partial);
    expect(MatchBand.fromScore(50), MatchBand.partial);
    expect(MatchBand.fromScore(49), MatchBand.weak);
    expect(MatchBand.fromScore(0), MatchBand.weak);
  });
}
