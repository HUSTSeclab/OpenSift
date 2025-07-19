# GitMetricsSync

## Installation

Ensure you have Go installed on your system. You will also need a PostgreSQL database configured to store the Git metrics.

1. Install dependencies:

   ```bash
   go mod tidy
   ```

## Usage

1. Compile the application:

   ```
   make
   ```

2. Run the synchronization process:

   ```
   ./bin/gitmetricsync -config config.json
   ```

   Replace `config.json` with the path to your database configuration file if it is located elsewhere.