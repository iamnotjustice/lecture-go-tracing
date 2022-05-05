# Mastery: Go Tracing Lecture example

## Prerequisites

We use Jaeger as the tracing backend here, this means we need to run it somehow. Let's do it with Docker!

Run this command: 

```
docker run \
  --rm \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest \
  --log-level=debug
```

This all-in-one image contains every part of jaeger we discussed, that means Collector, Query, UI, Agent and in-memory storage solution. Note that you also can run each of these components separately, but that's outside of the scope for now.

Once it starts, you can access the Jaeger UI at http://localhost:16686

## What we're trying to achieve?

This example is a simple distributed app which consists of two microservices: one calls the other to format some data. The first gets the data and it's type and then passes it to the formatter service which handles prettifying itself. The caller then prints the formatted text it gets as a response to the STDOUT and shuts down.

We need to add tracing to it which should work across API boundaries. This way we can track RPCs, check start\finish times and latency across spans inside a trace and so on. We will add span tags and logs which should help us when tracking issues.

## Running this example

Start with the Jaeger all-in-one container. Once you have it running, you can simply "go run" the formatter service.
Then run the "client" service with two parameters: first is data you want to format, second - it's type (json\xml).

After you run the client a few times check out Jaeger UI to find your traces.

# Note

Please note that our main goal for the code in this example is to be "easy" and "followable". I tried to simplify it as much as possible, so it uses stdlib as much as possible. It is by no means a "clean" code and you can (and should!) do better in real world applications!