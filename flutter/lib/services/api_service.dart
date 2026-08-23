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

  // Called on HTTP 401 before giving up. Returns true when a new access token
  // was obtained, in which case the request is replayed once. Set by
  // SessionNotifier so a long listening session survives access token expiry.
  static Future<bool> Function()? onRefreshNeeded;

  final client = http.Client();
  final baseUrl = AppConfig.apiBaseUrl;

  Future<T> request<T>({
    required String uri,
    HttpMethod httpMethod = HttpMethod.get,
    String? id,
    Map<String, dynamic>? data,
    Map<String, String>? queryParams,
    Map<String, String> headers = const {},
    T Function(dynamic)? parser,
    bool notifyOnUnauthorized = true,
    bool allowRefreshRetry = true,
  }) async {
    Uri url = Uri.parse('$baseUrl/$uri');

    if (id != null) {
      url = Uri.parse('$baseUrl/$uri/$id');
    }

    if (queryParams != null) {
      url = url.replace(queryParameters: queryParams);
    }

    if (kDebugMode) {
      print('${httpMethod.name.toUpperCase()} : $url');
    }

    final String? body = data != null ? jsonEncode(data) : null;

    // Reads the current access token on every attempt, so the replay that
    // follows a refresh picks up the renewed one.
    Future<http.Response> send() async {
      final String? token = await StorageService.get(StorageKey.token);
      final requestHeaders = {
        ...headers,
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        if (token != null) 'Authorization': 'Bearer $token',
      };

      try {
        switch (httpMethod) {
          case HttpMethod.post:
            return await client.post(url, body: body, headers: requestHeaders);
          case HttpMethod.put:
            return await client.put(url, body: body, headers: requestHeaders);
          case HttpMethod.delete:
            return await client.delete(url, headers: requestHeaders);
          default:
            return await client.get(url, headers: requestHeaders);
        }
      } on http.ClientException catch (e) {
        throw ApiException(httpStatus: 0, message: 'Erreur réseau: $e');
      } catch (e) {
        throw ApiException(httpStatus: 0, message: 'Erreur inattendue: $e');
      }
    }

    http.Response response = await send();

    // An expired access token is recoverable: renew once, then replay.
    if (response.statusCode == HttpStatus.unauthorized &&
        allowRefreshRetry &&
        ApiService.onRefreshNeeded != null) {
      final renewed = await ApiService.onRefreshNeeded!();
      if (renewed) response = await send();
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

  // POST binary data (audio chunks).
  Future<void> rawPost({
    required String uri,
    required Uint8List body,
    required String contentType,
    Map<String, String> headers = const {},
  }) async {
    final url = Uri.parse('$baseUrl/$uri');
    final token = await StorageService.get(StorageKey.token);
    final response = await client.post(
      url,
      body: body,
      headers: {
        ...headers,
        'Content-Type': contentType,
        if (token != null) 'Authorization': 'Bearer $token',
      },
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ApiException(
        httpStatus: response.statusCode,
        message: response.body,
      );
    }
  }
}
