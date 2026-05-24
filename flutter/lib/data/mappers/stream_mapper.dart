import '../models/stream_model.dart';
import '../../domain/entities/stream.dart';

/// Mapper to convert StreamModel (DTO) to StreamEntity (domain entity)
class StreamMapper {
  static StreamEntity toDomain(StreamModel model) {
    return StreamEntity(
      id: model.id,
      title: model.title,
      broadcasterId: model.broadcasterId,
      status: model.isLive ? StreamStatus.live : StreamStatus.ended,
      listenerCount: model.listenerCount,
    );
  }

  static List<StreamEntity> toDomainList(List<StreamModel> models) {
    return models.map(toDomain).toList();
  }
}
