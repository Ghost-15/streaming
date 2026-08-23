import 'package:integration_test/integration_test_driver.dart';

// Host-side driver: it collects binding.reportData from the device and writes it
// to build/integration_response_data.json, which is the measurement artefact.
Future<void> main() => integrationDriver();
