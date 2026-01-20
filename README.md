# Hanzo Functions

Serverless compute platform for event-driven workloads with GPU support.

## Overview

Hanzo Functions is a high-performance serverless platform that enables you to deploy and run functions at scale. With support for multiple runtimes, automatic scaling, and GPU acceleration, it is ideal for AI/ML workloads and event-driven architectures.

## Features

- **Auto-Scaling** - Scale from zero to thousands of instances automatically
- **GPU Support** - Native NVIDIA GPU support for AI/ML workloads
- **Multiple Runtimes** - Python, Go, Node.js, Rust, and custom containers
- **Event Sources** - HTTP, Kafka, RabbitMQ, Cron, and custom triggers
- **Cold Start Optimization** - Sub-100ms cold starts with prewarming
- **Versioning** - Blue/green deployments and canary releases
- **Observability** - Built-in metrics, logging, and tracing

## Quick Start

### Docker

```bash
docker run -d --name hanzo-functions \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  hanzoai/functions:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  functions:
    image: hanzoai/functions:latest
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - REGISTRY_URL=registry.local:5000

  registry:
    image: registry:2
    ports:
      - "5000:5000"
```

### Kubernetes (Helm)

```bash
helm repo add hanzo https://charts.hanzo.ai
helm install functions hanzo/functions
```

## CLI Installation

```bash
# macOS
brew install hanzoai/tap/hanzo-fn

# Linux
curl -sSL https://get.hanzo.ai/fn | bash

# npm
npm install -g @hanzo/fn
```

## Example Functions

### Python

```python
# handler.py
def handler(context, event):
    name = event.body.get("name", "World")
    return {
        "statusCode": 200,
        "body": f"Hello, {name}!"
    }
```

```bash
hanzo-fn deploy hello --runtime python:3.11 --handler handler.py
```

### Go

```go
// handler.go
package main

import "github.com/hanzoai/functions-go"

func Handler(ctx *functions.Context, event functions.Event) interface{} {
    return map[string]interface{}{
        "statusCode": 200,
        "body":       "Hello from Go!",
    }
}
```

### Node.js

```javascript
// handler.js
exports.handler = async (context, event) => {
    return {
        statusCode: 200,
        body: JSON.stringify({ message: "Hello from Node.js!" })
    };
};
```

### GPU Function (AI/ML)

```python
# inference.py
import torch
from transformers import pipeline

model = pipeline("text-generation", model="gpt2", device=0)

def handler(context, event):
    prompt = event.body.get("prompt", "Hello")
    result = model(prompt, max_length=100)
    return {"statusCode": 200, "body": result}
```

```bash
hanzo-fn deploy inference \
  --runtime python:3.11-cuda \
  --gpu 1 \
  --handler inference.py
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Hanzo Functions                      │
│                                                          │
│  ┌──────────┐   ┌──────────┐   ┌──────────────────────┐ │
│  │ Gateway  │──▶│ Executor │──▶│ Function Containers  │ │
│  └──────────┘   └──────────┘   │  ┌────┐ ┌────┐      │ │
│       │                        │  │ Fn │ │ Fn │ ...  │ │
│       │         ┌──────────┐   │  └────┘ └────┘      │ │
│       └────────▶│ Scaler   │   └──────────────────────┘ │
│                 └──────────┘                             │
│                      │                                   │
│              ┌───────┴───────┐                          │
│              │   Registry    │                          │
│              └───────────────┘                          │
└─────────────────────────────────────────────────────────┘
```

## Supported Runtimes

| Runtime | Versions | GPU Support |
|---------|----------|-------------|
| Python | 3.9, 3.10, 3.11, 3.12 | Yes |
| Go | 1.20, 1.21, 1.22 | Yes |
| Node.js | 18, 20, 22 | No |
| Rust | 1.75+ | Yes |
| Custom | Any Docker image | Yes |

## Documentation

- [Getting Started](https://docs.hanzo.ai/functions/getting-started)
- [Runtimes](https://docs.hanzo.ai/functions/runtimes)
- [GPU Functions](https://docs.hanzo.ai/functions/gpu)
- [Event Sources](https://docs.hanzo.ai/functions/events)
- [Deployment](https://docs.hanzo.ai/functions/deployment)

## License

MIT License - see [LICENSE](LICENSE) for details.
