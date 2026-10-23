import 'package:flutter_test/flutter_test.dart';

import 'package:gapcheck/main.dart';

void main() {
  testWidgets('home page shows the app name', (tester) async {
    await tester.pumpWidget(const GapCheckApp());

    expect(find.text('GapCheck'), findsOneWidget);
    expect(
      find.text('Score your resume against a job description.'),
      findsOneWidget,
    );
  });
}
