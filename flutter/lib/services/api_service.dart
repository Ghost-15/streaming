import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../helpers/exceptions.dart';
import 'storage_service.dart';

enum HttpMethod { get, post, put, delete }

class ApiService {
  static final ApiService _instance = ApiService._internal();

  factory ApiService() => _instance;

  ApiService._internal();

  // Called on HTTP 401 (except when notifyOnUnauthorized is false).
  static void Function()? onUnauthorized;

  final client = http.Client();
  final baseUrl = AppConfig.apiBaseUrl;

  Future<T> request<T>({
    required String uri,
    HttpMethod httpMethod = HttpMethod.get,
    String? id,
    Map<String, dynamic>? data,
    Map<String, String>? queryParams,
    T Function(dynamic)? parser,
    bool notifyOnUnauthorized = true,
  }) async {
    Uri url = Uri.parse('$baseUrl/$uri');

    if (id != null) {
      url = Uri.parse('$baseUrl/$uri/$id');
    }

    if (queryParams != null) {
      url = url.replace(queryParameters: queryParams);
    }

    final String? token = await StorageService.get(StorageKey.token);

    if (kDebugMode) {
      print('${httpMethod.name.toUpperCase()} : $url');
    }

    final headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };

    final String? body = data != null ? jsonEncode(data) : null;
    http.Response response;

    try {
      switch (httpMethod) {
        case HttpMethod.post:
          response = await client.post(url, body: body, headers: headers);
          break;
        case HttpMethod.put:
          response = await client.put(url, body: body, headers: headers);
          break;
        case HttpMethod.delete:
          response = await client.delete(url, headers: headers);
          break;
        default:
          response = await client.get(url, headers: headers);
      }
    } on http.ClientException catch (e) {
      throw ApiException(httpStatus: 0, message: 'Erreur réseau: $e');
    } catch (e) {
      throw ApiException(httpStatus: 0, message: 'Erreur inattendue: $e');
    }

    switch (response.statusCode) {
      case HttpStatus.created:
      case HttpStatus.ok:
        if (response.body.isEmpty) return null as T;
        final decoded = jsonDecode(response.body);
        if (parser != null) return parser(decoded);
        return decoded as T;
      case HttpStatus.noContent:
        return null as T;
      case HttpStatus.unauthorized:
        if (notifyOnUnauthorized) ApiService.onUnauthorized?.call();
        throw ApiException(httpStatus: 401, message: response.body);
      default:
        throw ApiException(
          httpStatus: response.statusCode,
          message: response.body,
        );
    }
  }

  // POST binary data (audio chunks). Best-effort: errors are silently dropped.
  Future<void> rawPost({
    required String uri,
    required Uint8List body,
    required String contentType,
  }) async {
    final url = Uri.parse('$baseUrl/$uri');
    final token = await StorageService.get(StorageKey.token);
    try {
      await client.post(url, body: body, headers: {
        'Content-Type': contentType,
        if (token != null) 'Authorization': 'Bearer $token',
      });
    } catch (_) {}
  }
}
