class ApiException implements Exception {
  final int httpStatus;
  final String message;

  const ApiException({required this.httpStatus, required this.message});

  @override
  String toString() => 'ApiException($httpStatus): $message';
}
