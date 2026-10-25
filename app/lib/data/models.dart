/// Dart mirrors of the API's JSON contract (backend/internal/analysis/types.go).
library;

/// Length limits for each text box. Same numbers as the backend's
/// MinTextLen / MaxTextLen, so bad input is caught before any request.
const minTextLength = 50;
const maxTextLength = 20000;

/// One concrete improvement for the resume.
class Suggestion {
  const Suggestion({
    required this.section,
    required this.issue,
    required this.fix,
  });

  final String section;
  final String issue;
  final String fix;

  factory Suggestion.fromJson(Map<String, dynamic> json) => Suggestion(
    section: json['section'] as String? ?? '',
    issue: json['issue'] as String? ?? '',
    fix: json['fix'] as String? ?? '',
  );
}

/// How strong the match is, using the same bands as the LLM's rules.
enum MatchBand {
  excellent('Excellent match'),
  strong('Strong match'),
  partial('Partial match'),
  weak('Weak match');

  const MatchBand(this.label);
  final String label;

  static MatchBand fromScore(int score) => switch (score) {
    >= 90 => MatchBand.excellent,
    >= 70 => MatchBand.strong,
    >= 50 => MatchBand.partial,
    _ => MatchBand.weak,
  };
}

/// What POST /api/analyze returns on success.
class AnalysisResult {
  const AnalysisResult({
    required this.score,
    required this.summary,
    required this.matchedKeywords,
    required this.missingKeywords,
    required this.suggestions,
  });

  final int score;
  final String summary;
  final List<String> matchedKeywords;
  final List<String> missingKeywords;
  final List<Suggestion> suggestions;

  MatchBand get band => MatchBand.fromScore(score);

  factory AnalysisResult.fromJson(Map<String, dynamic> json) => AnalysisResult(
    score: (json['score'] as num?)?.toInt() ?? 0,
    summary: json['summary'] as String? ?? '',
    matchedKeywords: _strings(json['matched_keywords']),
    missingKeywords: _strings(json['missing_keywords']),
    suggestions: [
      for (final s in json['suggestions'] as List? ?? const [])
        if (s is Map<String, dynamic>) Suggestion.fromJson(s),
    ],
  );

  static List<String> _strings(Object? value) => [
    for (final v in value as List? ?? const [])
      if (v is String) v,
  ];
}

/// An error from the API, or from failing to reach it.
///
/// [code] matches the backend's error codes (invalid_input, rate_limited,
/// timeout, upstream_error, ...) plus two client-side ones: `network` and
/// `unknown`.
class ApiException implements Exception {
  const ApiException({
    required this.code,
    required this.message,
    this.field,
    this.statusCode,
  });

  final String code;
  final String message;
  final String? field;
  final int? statusCode;

  @override
  String toString() => 'ApiException($code, $statusCode): $message';
}
