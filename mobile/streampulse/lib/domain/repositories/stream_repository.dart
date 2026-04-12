// Repository interface — domain layer.
import '../entities/stream.dart';

abstract interface class StreamRepository {
  Future<List<StreamEntity>> listActive();
  Future<StreamEntity> getById(String id);
  Future<StreamEntity> start({required String title});
  Future<void> end(String streamId);
}
