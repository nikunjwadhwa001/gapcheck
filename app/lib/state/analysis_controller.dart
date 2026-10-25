import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/api_client.dart';
import '../data/models.dart';

/// The API client. Tests replace it with
/// `apiProvider.overrideWithValue(GapCheckApi(..., client: MockClient(...)))`.
final apiProvider = Provider<GapCheckApi>(
  (ref) => GapCheckApi(baseUrl: apiBaseUrl),
);

/// The current analysis:
/// - `null` before the first request
/// - `AsyncLoading` while waiting
/// - `AsyncData(result)` on success
/// - `AsyncError(ApiException)` on failure
final analysisProvider =
    NotifierProvider<AnalysisController, AsyncValue<AnalysisResult>?>(
      AnalysisController.new,
    );

class AnalysisController extends Notifier<AsyncValue<AnalysisResult>?> {
  ({String resume, String jd})? _lastInput;

  @override
  AsyncValue<AnalysisResult>? build() => null;

  Future<void> analyze({required String resume, required String jd}) async {
    _lastInput = (resume: resume, jd: jd);
    state = const AsyncLoading();
    // guard turns a thrown ApiException into AsyncError instead of crashing.
    state = await AsyncValue.guard(
      () => ref.read(apiProvider).analyze(resumeText: resume, jdText: jd),
    );
  }

  /// Sends the last input again (used by the "Try again" button).
  Future<void> retry() async {
    final input = _lastInput;
    if (input != null) await analyze(resume: input.resume, jd: input.jd);
  }
}
