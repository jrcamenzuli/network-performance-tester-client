# Network Performance Tester (WORK IN PROGRESS)

The goal of this project is to test various network performance metrics automatically and produce CSV results. This test suite can be used to test the performance of the network itself, or a filtering application, for example. These tests can help find regressions or areas of improvement in what is being tested. Various measurements can be made such as: CPU usage, Memory usage, error rate, transfer rate and various timings.

## Testing Methodology

All tests require a minimum of two devices:

1. A DUT (device under test)
2. A server that is direct-connected, uncontested and dedicated to the purpose of performance testing.

Every attempt should be made to eliminate all variables that may introduce inconsistencies and reduce reproducibility of measurements. For example, WiFi should not be used to connect the DUT and server, due to the inherent nature of the shared radio spectrum.

## Types of tests

### Device Idle Test

This test is used to get a baseline measurement of performance for a system.

### Process Idle Test

This test is used to get a baseline measurement of performance for a OS process.

### HTTP Burst

A burst of HTTP requests made at increasing sizes between rest periods. The test stops when the device's limit has been reached or some predefined maximum burst size has been reached.

### HTTP Rate

HTTP requests made at increasing rates between rest periods. The test stops when the device's limit has been reached or some predefined maximum rate size has been reached.

### Throughput

The maximum throughput that is possible. Throughput will be measured in both directions independently and in both directions at the same time. 

### Ping

The measured Ping or RTT.

### Jitter

The measured Jitter.

### DNS Burst

A burst of DNS queries made at increasing sizes between rest periods. The test stops when the device's limit has been reached or some predefined maximum query burst size has been reached.

### DNS Rate

DNS queries made at increasing rates between rest periods. The test stops when the device's limit has been reached or some predefined maximum query rate size has been reached.