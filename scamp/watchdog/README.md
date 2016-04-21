Watchdog
===

The SCAMP watch dog keeps track of how many processes are running. When the count is falls below the chosen number the watch dog will kick off an event. The configuration format:

```
{
  "main": {
    "Thing.method~1": 10
  }
}
```

is meant to be concise and human editable.

Dumping Your Inventory
---

gcloud docker pull gcr.io/retailops-1/scamp-watchdog:dev && docker run -v $PWD/inventory.json:/inventory.json --rm -it --volumes-from=cache gcr.io/retailops-1/scamp-watchdog:dev watchdog -config /backplane/etc/soa.conf -expected-inventory /inventory.json -mode dump-inventory

Running The Watchdog Against The Inventory
---

gcloud docker pull gcr.io/retailops-1/scamp-watchdog:dev && docker run -v $PWD/inventory.json:/inventory.json --rm -it --volumes-from=cache gcr.io/retailops-1/scamp-watchdog:dev watchdog -config /backplane/etc/soa.conf -expected-inventory /inventory.json