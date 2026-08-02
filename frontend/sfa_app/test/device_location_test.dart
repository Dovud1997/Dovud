import 'package:flutter_test/flutter_test.dart';
import 'package:sfa_app/core/device/device_location.dart';

void main() {
  test('DevicePosition.toJson includes coords and source', () {
    const p = DevicePosition(lat: 41.3, lng: 69.2, accuracy: 12.5, source: 'device');
    expect(p.toJson(), {
      'lat': 41.3,
      'lng': 69.2,
      'accuracy': 12.5,
      'source': 'device',
    });
  });

  test('demoTashkent is marked as demo source', () {
    expect(DeviceLocationService.demoTashkent.source, 'demo');
    expect(DeviceLocationService.demoTashkent.lat, closeTo(41.3111, 0.0001));
  });

  test('currentOrDemo falls back when allowDemoFallback', () async {
    // Without a platform geolocator mock, current() fails in unit tests.
    final svc = DeviceLocationService();
    final pos = await svc.currentOrDemo(allowDemoFallback: true);
    expect(pos.source, 'demo');
  });
}
