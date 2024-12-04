# HTTP Monitor

A one-shot tool for monitoring HTTP and HTTPS endpoints, designed for simplicity and integration into existing workflows.

## Overview

**HTTP Monitor (httpmon)** is a lightweight tool for operations engineers to monitor HTTP endpoints. It checks critical metrics like response time, status, and TLS certificate validity, outputting results in CSV format to standard output (stdout). Use it with schedulers like `cron` for continuous monitoring.

## Key Features

- Monitor single or multiple HTTP/HTTPS endpoints.
- Load targets from a file (one URL per line).
- Customize monitor names for easier identification.
- Detailed performance metrics for HTTP requests.

## Quick Start

### Prerequisites

- Install Go if you plan to build the tool from source.

### Usage

1. **Monitor one or more URLs:**
   ```bash
   httpmon [URL]...
   ```

2. **Monitor URLs from a file (one URL per line):**
   ```bash
   httpmon -f [FILENAME]
   ```

3. **Set a custom monitor name:**
   ```bash
   httpmon -n my-monitor [URL]...
   ```

4. **Save results to a log file:**
   ```bash
   httpmon [URL]... >> monitoring.log
   ```

### Using with Cron for Continuous Monitoring

Schedule regular monitoring by combining `httpmon` with `cron`. For example, to run every 5 minutes and append results to `monitoring.log`:

```bash
*/5 * * * * /path/to/httpmon https://example.com >> /path/to/monitoring.log
```

### Output Format

The results are printed to stdout in CSV format with the following columns:

| Column Name               | Description                                   |
|---------------------------|-----------------------------------------------|
| **Monitor Name**          | Name assigned to the monitor.                |
| **URL**                   | The target URL being monitored.              |
| **Status**                | Monitoring result (e.g., success or failure).|
| **Timestamp**             | Time of the check (UTC).                     |
| **Status Code**           | HTTP status code (e.g., 200, 404).           |
| **Message**               | Additional status details.                   |
| **DNS Time (ms)**         | Time spent resolving DNS.                    |
| **Connection Time (ms)**  | Time taken to establish a connection.        |
| **TLS Time (ms)**         | Time spent establishing a TLS handshake.     |
| **TTFB (ms)**             | Time to first byte.                          |
| **Download Time (ms)**    | Time spent downloading the response.         |
| **Total Response Time (ms)** | Total time for the request.                |
| **Cert Validity (s)**     | Remaining validity of the TLS certificate.   |

### Examples

1. Monitor two URLs:
   ```bash
   httpmon https://example.com https://example.org
   ```

2. Monitor URLs from a file (`targets.txt` with one URL per line):
   ```bash
   httpmon -f targets.txt
   ```

3. Monitor a URL with a custom name:
   ```bash
   httpmon -n api-monitor https://api.example.com
   ```

4. Append output to a log file:
   ```bash
   httpmon https://example.com >> monitoring.log
   ```

## Who Should Use This Tool?

This tool is ideal for anyone who needs a simple, one-shot utility for gathering HTTP endpoint performance data. Pair it with `cron` or other schedulers for continuous monitoring and logging.
