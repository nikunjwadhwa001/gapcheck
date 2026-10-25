import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import 'models.dart';

/// Where the API lives. Set at build time:
///   flutter run --dart-define=API_BASE_URL=https://gapcheck-api.onrender.com
/// Defaults to the local Go server.
const apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080',
);

/// Talks to the GapCheck Go API.
class GapCheckApi {
  GapCheckApi({
    required this.baseUrl,
    http.Client? client,
    // Longer than the server's own 30s limit, with room for a free-tier
    // server waking up from sleep (Phase 3).
    this.timeout = const Duration(seconds: 90),
  }) : _client = client ?? http.Client();

  final String baseUrl;
  final Duration timeout;
  final http.Client _client;

  Future<AnalysisResult> analyze({
    required String resumeText,
    required String jdText,
  }) async {
    final http.Response res;
    try {
      res = await _client
          .post(
            Uri.parse('$baseUrl/api/analyze'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'resume_text': resumeText, 'jd_text': jdText}),
          )
          .timeout(timeout);
    } on TimeoutException {
      throw const ApiException(
        code: 'timeout',
        message: 'The server took too long to answer. Please try again.',
      );
    } on http.ClientException {
      // No response at all: offline, server down, or blocked by CORS.
      throw const ApiException(
        code: 'network',
        message: "Couldn't reach the GapCheck server. Check your connection and try again.",
      );
    }

    final body = _tryDecode(res.body);

    if (res.statusCode == 200 && body is Map<String, dynamic>) {
      return AnalysisResult.fromJson(body);
    }

    // Every API error looks like {"error": {"code", "message", "field?"}}.
    final error = body is Map<String, dynamic> ? body['error'] : null;
    if (error is Map<String, dynamic>) {
      throw ApiException(
        code: error['code'] as String? ?? 'unknown',
        message: error['message'] as String? ?? 'Something went wrong.',
        field: error['field'] as String?,
        statusCode: res.statusCode,
      );
    }
    throw ApiException(
      code: 'unknown',
      message: 'Unexpected response from the server (HTTP ${res.statusCode}).',
      statusCode: res.statusCode,
    );
  }

  /// Pings /health so a sleeping free-tier server starts waking up while
  /// the user is still pasting text. Failures don't matter here.
  Future<void> wakeUp() async {
    try {
      await _client.get(Uri.parse('$baseUrl/health')).timeout(timeout);
    } catch (_) {
      // The real request will report any problem properly.
    }
  }

  static Object? _tryDecode(String body) {
    try {
      return jsonDecode(body);
    } on FormatException {
      return null;
    }
  }
}
