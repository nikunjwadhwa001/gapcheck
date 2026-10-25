import 'package:go_router/go_router.dart';

import 'screens/input_screen.dart';
import 'screens/results_screen.dart';

/// Two screens: `/` to paste texts, `/results` to see the analysis.
///
/// Flutter web uses hash URLs by default (`/#/results`), which matters for
/// GitHub Pages: it only serves real files, so a refresh on a path like
/// `/results` would 404, but `/#/results` always loads index.html.
GoRouter createRouter() => GoRouter(
  routes: [
    GoRoute(path: '/', builder: (context, state) => const InputScreen()),
    GoRoute(
      path: '/results',
      builder: (context, state) => const ResultsScreen(),
    ),
  ],
);
