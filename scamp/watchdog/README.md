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