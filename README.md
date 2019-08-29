# grpc-sample


This sample app walks though setting [gRPC](https://grpc.io/) service on Cloud Run. This functionality is still experimental so some of the implementation may still change.

> Note, to keep this readme short, I will be asking you to execute scripts rather than listing here complete commands. You should really review each one of these scripts for content, and, to understand the individual commands so you can use them in the future.

## Pre-requirements

### GCP Project and gcloud SDK

If you don't have one already, start by creating new project and configuring [Google Cloud SDK](https://cloud.google.com/sdk/docs/). Similarly, if you have not done so already, you will have [set up Cloud Run](https://cloud.google.com/run/docs/setup).

## Setup

To setup this service you will need to clone this repo:

```shell
git clone https://github.com/mchmarny/logo-identifier.git
```

And navigate into that directory:

```shell
cd logo-identifier
```

## API Definition

First, you need to define the API, the shape of the payload that will be exchange between client and the server. Do define the Services and Messages you must use [Protocol Buffers](https://developers.google.com/protocol-buffers/) (Protobuf) which is a language-neutral mechanism for serializing structured data.

The `proto` file I will use in this example is located in [api/v1/message.proto](api/v1/message.proto) and it looks like this:

```protobuf
syntax = "proto3";

package message;

import "google/protobuf/timestamp.proto";
import "google/api/annotations.proto";

service MessageService {
  rpc Send(Request) returns (Response) {
    option (google.api.http) = {
      get: "/v1/send/{message}"
    };
  }
  rpc SendStream(stream Request) returns (stream Response) {}
}

message Request {
  string message = 1;
}

message Content {
  int32 index = 1;
  string message = 2;
  google.protobuf.Timestamp received_on = 3;
}

message Response {
  Content content = 1;
}
```

The content should be pretty self-explanatory. It basically describes the Request and Response, the Content message that both of these use, and the two methods: Send with unary request and response as well as SendStream with unary request and stream response. To learn more about setting up Protobuf see `go` support doc [here](https://github.com/golang/protobuf).


To auto-generate the `go` code from that `proto` run [bin/api](bin/api) script


```shell
bin/api
```

As a result, you should now have a new `go` files titled [pkg/api/v1/message.pb.go](pkg/api/v1/message.pb.go) and [pkg/api/v1/message.pb.gw.go](pkg/api/v1/message.pb.gw.go). You can review that file but don't edit it as it will be overwritten the next time we run the [bin/api](bin/api) script


## Container Image

Next, build the server container image which will be used to deploy Cloud Run service using the [bin/image](bin/image) script

```shell
bin/image
```

## Service Account

Now create a service account and assign it the necessary roles using the [bin/user](bin/user) script

```shell
bin/user
```

## Service Deployment

Once the container image and service account are ready, you can now deploy the new service using [bin/deploy](bin/deploy) script

```shell
bin/deploy
```

## Build Client

To invoke the deployed Cloud Run service, build the gRPC client using the [bin/client](bin/client) script

```shell
bin/client
```

The resulting CLI will be compiled into the `bin` directory. The output of the [bin/client](bin/client) script will also print out the two ways you can execute that client

```shell
Client CLI generated.
Usage:
 Unary Request/Unary Response
 bin/cli --server grpc-sample-***-uc.a.run.app:443 --message hi

 Unary Request/Stream Response
 bin/cli --server grpc-sample-***-uc.a.run.app:443 --message hi --stream 5
```

### Testing Service on Cloud Run

When executing the built CLI in unary way (by not including the `--stream` flag) you will see the details of the sent and received message

```shell
Unary Request/Unary Response
 Sent:
  hi
 Response:
  content:<index:1 message:"hi" received_on:<seconds:1567098976 nanos:535796117 > >
```

Where as executing it using stream (with `--stream` number) the CLI will print the sent message index and server processing time

```shell
Unary Request/Stream Response
  Stream[1] - Server time: 2019-08-29T17:16:22.837297811Z
  Stream[2] - Server time: 2019-08-29T17:16:22.837928885Z
  Stream[3] - Server time: 2019-08-29T17:16:22.83794915Z
  Stream[4] - Server time: 2019-08-29T17:16:22.837959711Z
  Stream[5] - Server time: 2019-08-29T17:16:22.837968925Z
```

> The gRPC service has also support for REST method `get: "/v1/send/{message}"` but this doesn't seem to work on Cloud Run right now. I'm still debugging this.

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.