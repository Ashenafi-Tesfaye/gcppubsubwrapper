# gcppubsubwrapper

A simple Go wrapper for the latest Google Cloud Pub/Sub client, providing a stable API for projects using older Pub/Sub versions.

## Features
- Simple publish and subscribe methods
- Uses the latest stable Google Pub/Sub Go client

## Usage

```go
package main

import (
	"context"
	"fmt"
	"gcppubsubwrapper"
)

func main() {
	ctx := context.Background()
	client, err := gcppubsubwrapper.NewPubSubClient(ctx, "your-gcp-project-id")
	if err != nil {
		panic(err)
	}

	// Publish example
	msgID, err := client.Publish(ctx, "your-topic-id", []byte("hello world"), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Published message ID:", msgID)

	// Subscribe example
	handler := func(ctx context.Context, msg *pubsub.Message) {
		fmt.Printf("Received: %s\n", string(msg.Data))
		msg.Ack()
	}
	// This will block and receive messages
	// err = client.Subscribe(ctx, "your-subscription-id", handler)
	// if err != nil {
	// 	panic(err)
	// }
}
```

## License
MIT
