import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sfa_app/core/offline/file_upload_queue.dart';
import 'package:sfa_app/features/documents/data/documents_repository.dart';

final documentsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.watch(documentsRepositoryProvider).listDocuments();
});

final pendingUploadsProvider = FutureProvider.autoDispose<List<PendingUpload>>((ref) {
  return ref.watch(fileUploadQueueProvider).list();
});

class DocumentsPage extends ConsumerStatefulWidget {
  const DocumentsPage({super.key});

  @override
  ConsumerState<DocumentsPage> createState() => _DocumentsPageState();
}

class _DocumentsPageState extends ConsumerState<DocumentsPage> {
  bool _busy = false;

  Future<void> _queueDemoUpload() async {
    setState(() => _busy = true);
    try {
      final bytes = utf8.encode('SFA offline upload ${DateTime.now().toIso8601String()}');
      await ref.read(fileUploadQueueProvider).enqueue(
            fileName: 'offline-note-${DateTime.now().millisecondsSinceEpoch}.txt',
            mime: 'text/plain',
            bytes: bytes,
          );
      ref.invalidate(pendingUploadsProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Queued offline upload')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _flushUploads() async {
    setState(() => _busy = true);
    try {
      final res = await ref.read(fileUploadQueueProvider).flush();
      ref.invalidate(pendingUploadsProvider);
      ref.invalidate(documentsProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Uploads: ${res['uploaded']} ok, ${res['failed']} failed')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('$e')));
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(documentsProvider);
    final pending = ref.watch(pendingUploadsProvider).valueOrNull ?? const [];
    return Scaffold(
      appBar: AppBar(title: const Text('Documents')),
      floatingActionButton: FloatingActionButton(
        onPressed: _busy ? null : _queueDemoUpload,
        child: const Icon(Icons.upload_file),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('Offline upload queue (${pending.length})',
              style: Theme.of(context).textTheme.titleMedium),
          if (pending.isEmpty)
            const ListTile(contentPadding: EdgeInsets.zero, title: Text('No pending uploads'))
          else
            ...pending.map(
              (u) => ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(u.fileName),
                subtitle: Text('${u.mime} · ${u.sizeBytes} bytes · ${u.status}'),
              ),
            ),
          Align(
            alignment: Alignment.centerLeft,
            child: FilledButton.tonal(
              onPressed: _busy || pending.isEmpty ? null : _flushUploads,
              child: const Text('Flush uploads'),
            ),
          ),
          const Divider(height: 32),
          Text('Documents', style: Theme.of(context).textTheme.titleMedium),
          async.when(
            data: (items) => items.isEmpty
                ? const ListTile(contentPadding: EdgeInsets.zero, title: Text('No documents'))
                : Column(
                    children: items
                        .map(
                          (d) => ListTile(
                            contentPadding: EdgeInsets.zero,
                            title: Text(d['title']?.toString() ?? ''),
                            subtitle: Text('${d['doc_type'] ?? ''} · ${d['status'] ?? ''}'),
                          ),
                        )
                        .toList(),
                  ),
            loading: () => const LinearProgressIndicator(),
            error: (e, _) => Text('$e'),
          ),
        ],
      ),
    );
  }
}
