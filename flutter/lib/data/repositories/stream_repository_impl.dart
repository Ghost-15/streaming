// Data layer — implements domain/repositories/stream_repository.dart
// Sprint 1 — US-003.
import '../../domain/entities/stream.dart';
import '../../domain/repositories/stream_repository.dart';
import '../datasources/api_client.dart';

class StreamRepositoryImpl implements StreamRepository {
  // ignore: unused_field — sera utilisé Sprint 1 (US-003)
  final ApiClient _client;
  const StreamRepositoryImpl(this._client);

  @override
  Future<List<StreamEntity>> listActive() async {
    // TODO Sprint 1: GET /api/v1/streams
    throw UnimplementedError('listActive not yet implemented');
  }

  @override
  Future<StreamEntity> getById(String id) async {
    // TODO Sprint 1: GET /api/v1/streams/:id
    throw UnimplementedError('getById not yet implemented');
  }

  @override
  Future<StreamEntity> start({required String title}) async {
    // TODO Sprint 1: POST /api/v1/streams (diffuseur only)
    throw UnimplementedError('start not yet implemented');
  }

  @override
  Future<void> end(String streamId) async {
    // TODO Sprint 1: DELETE /api/v1/streams/:id
  }
}
