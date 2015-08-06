Features
========
 - [x] Setup Go project
 - [x] TLS session setup
 - [x] cert verification
 - [x] Parse packet
 - [x] Generate packet
 - [x] Generate request
   - [x] Generate request header JSON
 - [x] Parse reply
 - [x] Parse request
   - [x] Route to action based on header JSON
 - [x] Generate reply
 - [ ] Verify TLS certificate with `/etc/authorized_services`
 - [x] Manage connection msgno
 - [ ] Parse service cache
 - [ ] Route RPC based on service cache
 - [x] Use go logging library
 - [ ] AuthZ service support
 - [ ] Chunk body to 128k
 - [ ] Reconnect logic (`scamp.Connection` connects with exponential backoff)
  - [ ] What to do if connection goes down during `Session` exchange?

Important Restructuring
===

 - [ ] Stream messages bodies
   - [ ] Session stream interface? `Reader`/`Writer` for bytes? Benefit: integration with patterns/helpers in (io lib)[http://golang.org/pkg/io]
 - [ ] Unify concepts of `Request`/`Reply` with `Message` and move that distinction to the direction of the `Session`
   - [ ] Rewrite `Request`/`Reply` code to reuse `session` `Reader`/`Writer` under the hood
 
Rad Ideas
===

 - [ ] Ragel state machine specification to generate go code

Bugs
====
 - [ ] Fix bug where sending envelope type `JSON` fails silently (should at least emit 'unknown type' to STDERR)
 - [ ] Fix bug where header `"type": "request"` fails silently (should at least emit 'unknown type' to STDERR)
 - [ ] Fix reference to documentation `message_id` which should read `request_id`
 - [ ] Move to interface design. Message parts which implement `Packet` so we can specialize `Header` vs `Data` which have different bodies from different data types.