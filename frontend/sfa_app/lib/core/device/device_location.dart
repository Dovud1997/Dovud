import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';

final deviceLocationProvider = Provider<DeviceLocationService>((ref) {
  return DeviceLocationService();
});

class DevicePosition {
  const DevicePosition({
    required this.lat,
    required this.lng,
    this.accuracy,
    this.source = 'device',
  });

  final double lat;
  final double lng;
  final double? accuracy;
  final String source;

  Map<String, dynamic> toJson() => {
        'lat': lat,
        'lng': lng,
        if (accuracy != null) 'accuracy': accuracy,
        'source': source,
      };
}

/// Reads the device GPS (or browser geolocation on web).
class DeviceLocationService {
  /// Demo fallback used only when [allowDemoFallback] is true (tests / explicit).
  static const demoTashkent = DevicePosition(
    lat: 41.3111,
    lng: 69.2797,
    accuracy: 25,
    source: 'demo',
  );

  Future<bool> ensurePermission() async {
    final enabled = await Geolocator.isLocationServiceEnabled();
    if (!enabled) {
      throw StateError('Location services are disabled');
    }
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied) {
      throw StateError('Location permission denied');
    }
    if (permission == LocationPermission.deniedForever) {
      throw StateError('Location permission permanently denied — enable in settings');
    }
    return true;
  }

  /// Current position. Throws if permission / services unavailable.
  Future<DevicePosition> current({
    Duration timeLimit = const Duration(seconds: 20),
  }) async {
    await ensurePermission();
    final pos = await Geolocator.getCurrentPosition(
      locationSettings: LocationSettings(
        accuracy: LocationAccuracy.high,
        timeLimit: timeLimit,
      ),
    );
    return DevicePosition(
      lat: pos.latitude,
      lng: pos.longitude,
      accuracy: pos.accuracy.isFinite ? pos.accuracy : null,
      source: 'device',
    );
  }

  /// Best-effort: device GPS, or [demoTashkent] when [allowDemoFallback].
  Future<DevicePosition> currentOrDemo({bool allowDemoFallback = false}) async {
    try {
      return await current();
    } catch (_) {
      if (allowDemoFallback) return demoTashkent;
      rethrow;
    }
  }
}
