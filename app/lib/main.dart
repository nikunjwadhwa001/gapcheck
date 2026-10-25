import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'router.dart';
import 'theme.dart';

void main() {
  // ProviderScope holds the state of every Riverpod provider in the app.
  runApp(const ProviderScope(child: GapCheckApp()));
}

class GapCheckApp extends StatefulWidget {
  const GapCheckApp({super.key});

  @override
  State<GapCheckApp> createState() => _GapCheckAppState();
}

class _GapCheckAppState extends State<GapCheckApp> {
  // Created once per app, so navigation state survives rebuilds.
  final GoRouter _router = createRouter();

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'GapCheck',
      debugShowCheckedModeBanner: false,
      theme: buildTheme(Brightness.light),
      darkTheme: buildTheme(Brightness.dark),
      routerConfig: _router,
    );
  }
}
