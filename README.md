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

package ping;

import "google/protobuf/timestamp.proto";

service MessageService {
  rpc Send(Request) returns (Response) {}
  rpc SendStream(stream Request) returns (stream Response) {}
}

message Content {
  string body = 1;
  string author = 2;
  google.protobuf.Timestamp created_on = 3;
}

message Request {
  Content content = 1;
}

message Response {
  int32 index = 1;
  Content content = 2;
  google.protobuf.Timestamp received_on = 3;
}
```

The content should be pretty self-explanatory. It basically describes the Request and Response, the Content message that both of these use, and the two methods: Send with unary request and response as well as SendStream with unary request and stream response. To learn more about setting up Protobuf see `go` support doc [here](https://github.com/golang/protobuf).


To auto-generate the `go` code from that `proto` run [bin/api](bin/api) script


```shell
bin/api
```

As a result, you should now have a new `go` file titled [pkg/api/v1/message.pb.go](pkg/api/v1/message.pb.go). You can review that file but don't edit it as it will be overwritten the next time we run the [bin/api](bin/api) script


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
Client CLI geenrated.
Usage:
 Unary Request/Unary Response
 bin/cli --server grpc-sample-***-uc.a.run.app:443 --author username --message hi

 Unary Request/Stream Response
 bin/cli --server grpc-sample-***-uc.a.run.app:443 --author username --message hi --stream 5
```

### Testing Service on Cloud Run

When executing the built CLI in unary way (by not including the `--stream` flag) you will see the details of the sent and received message

```shell
Unary Request/Unary Response
 Sent:
  body:"hi" author:"mchmarny" created_on:<seconds:1567051011 nanos:881391000 >
 Response:
  index:1 content:<body:"hi" author:"mchmarny" created_on:<seconds:1567051011 nanos:881391000 > > received_on:<seconds:1567051012 nanos:294123569 >
```

Where as executing it using stream (with `--stream` number) the CLI will print the sent message index and server processing time

```shell
Unary Request/Stream Response
  Stream[1] - Server time: 2019-08-29T03:57:05.120052543Z
  Stream[2] - Server time: 2019-08-29T03:57:05.120096926Z
  Stream[3] - Server time: 2019-08-29T03:57:05.120117664Z
  Stream[4] - Server time: 2019-08-29T03:57:05.120130517Z
```

## Disclaimer

This is my personal project and it does not represent my employer. I take no responsibility for issues caused by this code. I do my best to ensure that everything works, but if something goes wrong, my apologies is all you will get.