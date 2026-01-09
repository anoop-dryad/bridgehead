# 🛰️ Bridgehead
**Unified Command Delivery & Intelligent Device Routing**

## 📖 Overview
**Bridgehead** is a high-performance communication layer designed to bridge the gap between cloud services and remote hardware. It provides a reliable, automated pathway for sending **downlink commands** to devices, regardless of the entry point or protocol.

The core strength of Bridgehead is its **Autonomous Routing Engine**, which dynamically resolves the correct path and gateway needed to reach a specific target device. By centralizing this logic, Bridgehead ensures that developers only need to specify *what* the device should do, while the system handles *how* to get the message there.

---

## 🏗️ Architecture & Deployment
Bridgehead is built as a **Multi-Entrypoint Monorepo** in Go, designed for cloud-native scalability. The project produces four specialized binaries deployed as parallel, independent pods:

| Service | Primary Function | Communication Pattern |
| :--- | :--- | :--- |
| **Rest-API** | Management & Authorization | HTTP/JSON |
| **MQTT Worker** | Device-level connectivity | Pub/Sub |
| **SQS Consumer** | Asynchronous task processing | Message Queues |
| **Kinesis Worker** | High-throughput state tracking | Data Streams |

All four services share a unified **Business Logic Layer** and **Data Schema** located in the `internal/` package, ensuring that a command sent via API is tracked and processed identically to one triggered by a stream event.

---

## 🚀 Key Features
* **Intelligent Path Resolution:** Automatically identifies and routes commands through the appropriate router based on real-time device mapping.
* **Protocol Agnostic:** Seamlessly handles interactions across REST, MQTT, SQS, and Kinesis.
* **Unified State Tracking:** Maintains a single source of truth for command lifecycles (Pending → Dispatched → Acknowledged).
* **Secure-by-Design:** Centralized authorization for user-facing entry points without impacting internal worker performance.

---

## 🛠️ Project Structure
```text
bridgehead/
├── cmd/                # Deployment Entry Points (4 Pods)
│   ├── rest-api/       # HTTP Server with Auth
│   ├── mqtt-worker/    # Device Pub/Sub logic
│   ├── sqs-worker/     # Queue Consumer
│   └── kinesis-worker/ # Stream Processor
├── internal/           # Private Shared Logic
│   ├── auth/           # Authorization & JWT (REST only)
│   ├── models/         # Shared Database Tables & Structs
│   ├── repository/     # Shared Data Access Logic
│   └── service/        # Core Downlink & Routing Logic
├── migrations/         # Database Schema (Atlas/SQL)
└── Makefile            # Multi-binary Build Management
