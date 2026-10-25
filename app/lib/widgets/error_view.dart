import 'package:flutter/material.dart';

import '../data/models.dart';

/// Turns an API error code into a title, icon and the right next step.
class ErrorView extends StatelessWidget {
  const ErrorView({
    super.key,
    required this.error,
    required this.onRetry,
    required this.onEdit,
  });

  final Object error;
  final VoidCallback onRetry;
  final VoidCallback onEdit;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final e = error is ApiException
        ? error as ApiException
        : const ApiException(code: 'unknown', message: 'Something went wrong.');
    final info = _infoFor(e.code);

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(info.icon, size: 48, color: theme.colorScheme.error),
              const SizedBox(height: 16),
              Text(
                info.title,
                style: theme.textTheme.titleLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              // The server's message is written for users, so show it as is.
              Text(
                e.message,
                style: theme.textTheme.bodyMedium,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 24),
              Wrap(
                spacing: 12,
                runSpacing: 8,
                alignment: WrapAlignment.center,
                children: [
                  if (info.canRetry)
                    FilledButton(
                      onPressed: onRetry,
                      child: const Text('Try again'),
                    ),
                  OutlinedButton(
                    onPressed: onEdit,
                    child: const Text('Edit inputs'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  static ({String title, IconData icon, bool canRetry}) _infoFor(String code) =>
      switch (code) {
        'network' => (
          title: "Can't reach the server",
          icon: Icons.wifi_off,
          canRetry: true,
        ),
        'timeout' => (
          title: 'That took too long',
          icon: Icons.hourglass_bottom,
          canRetry: true,
        ),
        'rate_limited' => (
          title: 'Too many requests',
          icon: Icons.speed,
          canRetry: true,
        ),
        'invalid_input' || 'invalid_json' || 'body_too_large' => (
          title: 'Check your input',
          icon: Icons.edit_note,
          canRetry: false,
        ),
        _ => (
          title: 'Something went wrong',
          icon: Icons.error_outline,
          canRetry: true,
        ),
      };
}
