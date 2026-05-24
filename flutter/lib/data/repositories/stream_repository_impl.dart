// Data layer — implements domain/repositories/stream_repository.dart
// Sprint 2 — US-009 (Audio streaming).
import '../../core/services/api_client.dart';
import '../../core/constants/app_constants.dart';
import '../../domain/entities/stream.dart';
import '../../domain/repositories/stream_repository.dart';
import '../models/stream_model.dart';
import '../mappers/stream_mapper.dart';

class StreamRepositoryImpl implements StreamRepository {
  final ApiClient _apiClient;

  StreamRepositoryImpl({ApiClient? apiClient})
      : _apiClient = apiClient ?? ApiClient();

  /// Get active streams from API
  @override
  Future<List<StreamEntity>> listActive() async {
    try {
      final response = await _apiClient.get<List<dynamic>>(
        AppConstants.streamsActiveEndpoint,
      );

      if (response.data == null) return [];

      final streams = (response.data as List)
          .map((json) => StreamModel.fromJson(json as Map<String, dynamic>))
          .toList();

      return StreamMapper.toDomainList(streams);
    } catch (e) {
      throw Exception('Failed to fetch active streams: $e');
    }
  }

  /// Get stream by ID
  @override
  Future<StreamEntity> getById(String id) async {
    try {
      final endpoint = AppConstants.streamDetailEndpoint.replaceFirst('{id}', id);
      final response = await _apiClient.get<Map<String, dynamic>>(endpoint);

      if (response.data == null) {
        throw Exception('Stream not found');
      }

      final streamModel = StreamModel.fromJson(response.data!);
      return StreamMapper.toDomain(streamModel);
    } catch (e) {
      throw Exception('Failed to fetch stream: $e');
    }
  }

  /// Start a new stream (broadcaster only)
  @override
  Future<StreamEntity> start({required String title}) async {
    try {
      final response = await _apiClient.post<Map<String, dynamic>>(
        AppConstants.streamsActiveEndpoint,
        data: {'title': title},
      );

      if (response.data == null) {
        throw Exception('Failed to start stream');
      }

      final streamModel = StreamModel.fromJson(response.data!);
      return StreamMapper.toDomain(streamModel);
    } catch (e) {
      throw Exception('Failed to start stream: $e');
    }
  }

  /// End a stream
  @override
  Future<void> end(String streamId) async {
    try {
      final endpoint = AppConstants.streamsActiveEndpoint + '/$streamId';
      await _apiClient.delete(endpoint);
    } catch (e) {
      throw Exception('Failed to end stream: $e');
    }
  }

  /// Get active streams (for ViewModel compatibility)
  Future<List<StreamEntity>> getActiveStreams() async {
    return listActive();
  }

  /// Get stream by ID (for ViewModel compatibility)
  Future<StreamEntity> getStreamById(String id) async {
    return getById(id);
  }
}
