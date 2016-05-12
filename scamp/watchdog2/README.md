Watchdog2
====

Scans the discovery cache and takes inventory of available services. This inventory is compared with the expected inventory and compared against two thresholds: yellow and red. Red and yellow correspond to different pagerduty API tokens.

Known issues:

  - The watchdog pings pagerduty with updates to the ticket on every scan during an incident. Updates to degraded actions are not sent to pagerduty once the initial description has been sent. If a system becomes more degraded after the initial alert we will not become aware through slack. We could solve this by creating multiple incidents and tracking them separately. The service degradation from Yellow to Red does work properly -- the Yellow incident is closed and the Red is opened.
  - Updates to degraded actions are compared with the last known good inventory. When the host is fixed a new service will come up but with a different ID. The watchdog will count the new service but still count the old service as missing. Minor detail that can be confusing to those who are looking at alerts.
  - Many services are missing a version (as evidenced from plaintext examination of the discovery cache). Their versions are set to "~0" in the tool.
  - The scamp library used by the watchdog does not have consistent support for processing data off io.Readers, most API calls expect filepaths and will open files. Those API calls are difficult to stub in tests and thus the test suit for watchdog2 is limited. The scamp library calls which take filepaths should be split in to fileopening and stream processing calls, then the watchdog test suit could be expanded.