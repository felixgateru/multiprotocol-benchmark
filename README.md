# Multiprotocol Benchmarking Tool

A comprehensive benchmarking tool for testing IoT protocol performance across multiple transports.

## Supported Protocols

- **MQTT**
- **CoAP**
- **HTTP**
- **WebSocket**

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [Releases page](https://github.com/felixgateru/multiprotocol-benchmark/releases).

### Build from Source

```bash
git clone https://github.com/felixgateru/multiprotocol-benchmark.git
cd multiprotocol-benchmark
make build
```

Or manually with Go:

```bash
go build -o multiprotocol-benchmark
```

### Install via Go

```bash
go install github.com/felixgateru/multiprotocol-benchmark@latest
```

## Configuration

The tool requires a `.env` file for configuration.

### Quick Start

1. **Copy the example configuration:**

   ```bash
   cp .env.example .env
   ```

   Or use the Makefile:

   ```bash
   make setup
   ```

2. **Edit `.env` with your credentials:**

   ```bash
   USERNAME=your_username
   PASSWORD=your_password
   ```

3. **Run the benchmark:**

   ```bash
   make run
   ```

   Or run the binary directly:

   ```bash
   ./multiprotocol-benchmark
   ```

### Configuration Options

#### Required Variables

- `USERNAME` - Your SuperMQ account username
- `PASSWORD` - Your SuperMQ account password

#### Optional Variables

- `DOMAIN_ID` - Existing domain ID (if empty, a new domain will be created)
- `CLIENT_COUNT` - Number of clients to create (default: 1)
- `CHANNEL_COUNT` - Number of channels to create (default: 1)
- `PROVISION_FILE` - Path to provision.toml file with pre-configured channels/clients

#### Protocol-Specific Settings

Each protocol can be enabled/disabled and configured independently:

**MQTT:**

- `RUN_MQTT` - Enable MQTT tests (default: true)
- `MQTT_BROKER` - MQTT broker address (default: tcp://localhost:1883)
- `MQTT_MESSAGE_COUNT` - Number of messages to send (default: 100)
- `MQTT_MESSAGE_SIZE` - Message size in bytes (default: 256)
- `MQTT_DELAY` - Delay between messages (default: 100ms)
- `MQTT_QOS` - Quality of Service level (0, 1, or 2)

**CoAP:**

- `RUN_COAP` - Enable CoAP tests
- `COAP_HOST` - CoAP server host
- `COAP_PORT` - CoAP server port (default: 5683)
- `COAP_MESSAGE_COUNT`, `COAP_MESSAGE_SIZE`, `COAP_DELAY`

**HTTP:**

- `RUN_HTTP` - Enable HTTP tests
- `HTTP_BASE_URL` - HTTP API base URL
- `HTTP_MESSAGE_COUNT`, `HTTP_MESSAGE_SIZE`, `HTTP_DELAY`

**WebSocket:**

- `RUN_WS` - Enable WebSocket tests
- `WS_BASE_URL` - WebSocket server URL
- `WS_MESSAGE_COUNT`, `WS_MESSAGE_SIZE`, `WS_DELAY`

#### Other Settings

- `TLS_VERIFY` - Enable TLS verification (default: false)
- `CA_CERT_PATH` - Path to custom CA certificate
- `SAVE_TO_FILE` - Save results to JSON file (default: true)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)

## Usage Examples

### Using Make Commands

```bash
# Build the project
make build

# Build and run
make run

# Run tests
make test

# Clean build artifacts
make clean

# Cross-compile for multiple platforms
make cross-compile

# Format code
make format

# View all available commands
make help
```

### Test Only MQTT Protocol

```bash
RUN_MQTT=true RUN_COAP=false RUN_HTTP=false RUN_WS=false ./multiprotocol-benchmark
```

### High-Performance Benchmark (No Delay)

```bash
MQTT_DELAY=0 COAP_DELAY=0 HTTP_DELAY=0 WS_DELAY=0 ./multiprotocol-benchmark
```

### Using Custom Provision File

If you have existing channels and clients, create a `provision.toml`:

```toml
channels = ["channel_id_1", "channel_id_2"]

[[clients]]
id = "client_id_1"
secret = "client_secret_1"

[[clients]]
id = "client_id_2"
secret = "client_secret_2"
```

Then run:

```bash
PROVISION_FILE=provision.toml ./multiprotocol-benchmark
```

## Output

The tool generates performance metrics including:

- Success/failure rates
- Message throughput
- Latency statistics
- Protocol-specific metrics

Results are displayed in the console and optionally saved to a timestamped JSON file (e.g., `test_results_20250127_143022.json`).
