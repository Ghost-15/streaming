// Use Case — Sprint 1 — US-003.
import '../entities/stream.dart';
import '../repositories/stream_repository.dart';

class ListStreamsUseCase {
  final StreamRepository _repo;
  const ListStreamsUseCase(this._repo);

  Future<List<StreamEntity>> call() => _repo.listActive();
}
