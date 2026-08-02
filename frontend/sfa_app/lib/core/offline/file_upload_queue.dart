import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/drift/sfa_database.dart';
import 'package:sfa_app/core/offline/drift/shared_database.dart';
import 'package:sfa_app/features/documents/data/documents_repository.dart';

typedef BytesUploader = Future<Map<String, dynamic>> Function({
  required String fileName,
  required String mime,
  required List<int> bytes,
});

final fileUploadQueueProvider = Provider<FileUploadQueue>((ref) {
  final docs = ref.watch(documentsRepositoryProvider);
  return FileUploadQueue(
    sharedSfaDatabase(),
    uploadBytes: docs.uploadBytes,
  );
});

class PendingUpload {
  PendingUpload({
    required this.uploadId,
    required this.fileName,
    required this.mime,
    required this.sizeBytes,
    required this.status,
    this.localPath,
    this.remoteFileId,
    this.error,
    required this.createdAt,
    this.bytes,
  });

  final String uploadId;
  final String fileName;
  final String mime;
  final int sizeBytes;
  final String status;
  final String? localPath;
  final String? remoteFileId;
  final String? error;
  final DateTime createdAt;
  final List<int>? bytes;
}

/// Offline-capable file upload queue backed by Drift `file_uploads`.
///
/// Payload bytes are stored in the `payload` blob column so pending uploads
/// survive process restarts (not only an in-memory map).
class FileUploadQueue {
  FileUploadQueue(this._db, {required BytesUploader uploadBytes})
      : _uploadBytes = uploadBytes;

  final SfaDatabase _db;
  final BytesUploader _uploadBytes;

  /// Hot cache; source of truth is Drift `payload`.
  final Map<String, List<int>> _bytes = {};

  Future<PendingUpload> enqueue({
    required String fileName,
    required String mime,
    required List<int> bytes,
    String? localPath,
  }) async {
    final id = 'up-${DateTime.now().microsecondsSinceEpoch}';
    final payload = Uint8List.fromList(bytes);
    _bytes[id] = payload;
    await _db.into(_db.fileUploads).insert(
          FileUploadsCompanion.insert(
            uploadId: id,
            fileName: fileName,
            mime: mime,
            sizeBytes: Value(bytes.length),
            localPath: Value(localPath),
            status: const Value('pending'),
            createdAt: DateTime.now().toUtc(),
            payload: Value(payload),
          ),
        );
    return PendingUpload(
      uploadId: id,
      fileName: fileName,
      mime: mime,
      sizeBytes: bytes.length,
      status: 'pending',
      localPath: localPath,
      createdAt: DateTime.now().toUtc(),
      bytes: payload,
    );
  }

  Future<List<PendingUpload>> list({String status = 'pending'}) async {
    final rows = await (_db.select(_db.fileUploads)
          ..where((t) => t.status.equals(status))
          ..orderBy([(t) => OrderingTerm.asc(t.createdAt)]))
        .get();
    return rows.map(_fromRow).toList();
  }

  PendingUpload _fromRow(FileUploadRow r) {
    final stored = r.payload;
    List<int>? bytes = _bytes[r.uploadId];
    if ((bytes == null || bytes.isEmpty) && stored != null && stored.isNotEmpty) {
      bytes = List<int>.from(stored);
      _bytes[r.uploadId] = bytes;
    }
    return PendingUpload(
      uploadId: r.uploadId,
      fileName: r.fileName,
      mime: r.mime,
      sizeBytes: r.sizeBytes,
      status: r.status,
      localPath: r.localPath,
      remoteFileId: r.remoteFileId,
      error: r.error,
      createdAt: r.createdAt.toUtc(),
      bytes: bytes,
    );
  }

  Future<List<int>?> _loadBytes(String uploadId) async {
    final cached = _bytes[uploadId];
    if (cached != null && cached.isNotEmpty) return cached;
    final row = await (_db.select(_db.fileUploads)
          ..where((t) => t.uploadId.equals(uploadId)))
        .getSingleOrNull();
    final stored = row?.payload;
    if (stored == null || stored.isEmpty) return null;
    final bytes = List<int>.from(stored);
    _bytes[uploadId] = bytes;
    return bytes;
  }

  Future<Map<String, dynamic>> flush() async {
    final pending = await list(status: 'pending');
    var uploaded = 0;
    var failed = 0;
    for (final item in pending) {
      final bytes = await _loadBytes(item.uploadId);
      if (bytes == null || bytes.isEmpty) {
        await _mark(item.uploadId, status: 'failed', error: 'missing bytes');
        failed++;
        continue;
      }
      try {
        final file = await _uploadBytes(
          fileName: item.fileName,
          mime: item.mime,
          bytes: bytes,
        );
        await _mark(
          item.uploadId,
          status: 'uploaded',
          remoteFileId: file['id']?.toString(),
          clearPayload: true,
        );
        _bytes.remove(item.uploadId);
        uploaded++;
      } catch (e) {
        await _mark(item.uploadId, status: 'failed', error: e.toString());
        failed++;
      }
    }
    return {
      'pending': pending.length,
      'uploaded': uploaded,
      'failed': failed,
      'remaining': (await list(status: 'pending')).length,
    };
  }

  Future<void> _mark(
    String uploadId, {
    required String status,
    String? remoteFileId,
    String? error,
    bool clearPayload = false,
  }) {
    return (_db.update(_db.fileUploads)..where((t) => t.uploadId.equals(uploadId))).write(
      FileUploadsCompanion(
        status: Value(status),
        remoteFileId: Value(remoteFileId),
        error: Value(error),
        payload: clearPayload ? const Value(null) : const Value.absent(),
      ),
    );
  }
}
