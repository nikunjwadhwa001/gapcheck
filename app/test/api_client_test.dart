import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:gapcheck/data/api_client.dart';
import 'package:gapcheck/data/models.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// Builds an API client whose HTTP calls go to [handler] instead of a server.
GapCheckApi apiWith(
  Future<http.Response> Function(http.Request) handler, {
  Duration timeout = const Duration(seconds: 5),
}) => GapCheckApi(
  baseUrl: 'http://api.test',
  client: MockClient(handler),
  timeout: timeout,
);

Future<AnalysisResult> analyze(GapCheckApi api) =>
    api.analyze(resumeText: 'resume', jdText: 'job');

Matcher apiError(String code, {String? field}) => isA<ApiException>()
    .having((e) => e.code, 'code', code)
    .having((e) => e.field, 'field', field);

void main() {
  test('sends the right request and parses a success', () async {
    late http.Request sent;
    final api = apiWith((req) async {
      sent = req;
      return http.Response(
        jsonEncode({
          'score': 81,
          'summary': 'Good.',
          'matched_keywords': ['Go'],
          'missing_keywords': [],
          'suggestions': [],
        }),
        200,
      );
    });

    final result = await analyze(api);

    expect(sent.method, 'POST');
    expect(sent.url.toString(), 'http://api.test/api/analyze');
    expect(sent.headers['Content-Type'], startsWith('application/json'));
    expect(jsonDecode(sent.body), {'resume_text': 'resume', 'jd_text': 'job'});
    expect(result.score, 81);
  });

  test('turns an API error body into ApiException with code and field', () {
    final api = apiWith(
      (_) async => http.Response(
        jsonEncode({
          'error': {
            'code': 'invalid_input',
            'message': 'resume_text is too short',
            'field': 'resume_text',
          },
        }),
        400,
      ),
    );

    expect(
      analyze(api),
      throwsA(apiError('invalid_input', field: 'resume_text')),
    );
  });

  test('rate limit keeps the server code', () {
    final api = apiWith(
      (_) async => http.Response(
        '{"error":{"code":"rate_limited","message":"Wait a minute."}}',
        429,
      ),
    );
    expect(analyze(api), throwsA(apiError('rate_limited')));
  });

  test('a non-JSON error page becomes code "unknown"', () {
    final api = apiWith(
      (_) async => http.Response('<html>Bad Gateway</html>', 502),
    );
    expect(analyze(api), throwsA(apiError('unknown')));
  });

  test('no response at all becomes code "network"', () {
    final api = apiWith((_) async => throw http.ClientException('refused'));
    expect(analyze(api), throwsA(apiError('network')));
  });

  test('wakeUp pings /health and never throws', () async {
    final paths = <String>[];
    await apiWith((req) async {
      paths.add('${req.method} ${req.url.path}');
      throw http.ClientException('still asleep');
    }).wakeUp();

    expect(paths, ['GET /health']);
  });

  test('a slow server becomes code "timeout"', () {
    final api = apiWith((_) async {
      await Future<void>.delayed(const Duration(milliseconds: 200));
      return http.Response('{}', 200);
    }, timeout: const Duration(milliseconds: 20));
    expect(analyze(api), throwsA(apiError('timeout')));
  });
}
