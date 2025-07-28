# OpenBYMAData Debug Mode

## Overview

The OpenBYMAData library includes comprehensive debug functionality to help developers:
- Verify field mappings are correct
- Detect when BYMA API field names change
- Troubleshoot data parsing issues
- Understand API response structures
- Monitor for API breaking changes

## Configuration

### Environment Variable

Create a `.env` file in your project root:

```env
# Set to true to enable debug logging
DEBUG=true
```

### Logger Implementation

You need to provide a logger that implements the `Logger` interface:

```go
type Logger interface {
    Debug(msg string, fields ...LogField)
    Info(msg string, fields ...LogField)
    Warn(msg string, fields ...LogField)
    Error(msg string, fields ...LogField)
}
```

## Debug Output

When debug mode is enabled, you'll see:

### 1. **HTTP Request/Response Logging**
```
🐛 DEBUG: Making request
   method: POST
   url: https://open.bymadata.com.ar/vanoms-be-core/rest/api/bymadata/free/index-price

🐛 DEBUG: Request completed
   status_code: 200
   response_size: 4921
```

### 2. **Raw API Response Data**
```
🐛 DEBUG: Raw API Response
   endpoint: index-price
   raw_json: {"data":[{"symbol":"G","description":"S&P BYMA Indice General"...
   parsed_structure: map[data:[map[description:S&P BYMA Indice General...]]]
```

### 3. **Field Names and Structure**
```
🐛 DEBUG: Available fields in first data item
   endpoint: index-price
   fields: [symbol, description, price, variation, highValue, minValue, previousClosingPrice]
   first_item: map[description:S&P BYMA Indice General symbol:G price:9.4611e+07...]
```

## Example Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/carvalab/openbymadata"
)

type MyLogger struct{}

func (l *MyLogger) Debug(msg string, fields ...openbymadata.LogField) {
    fmt.Printf("DEBUG: %s\n", msg)
    for _, field := range fields {
        fmt.Printf("  %s: %v\n", field.Key, field.Value)
    }
}

func (l *MyLogger) Info(msg string, fields ...openbymadata.LogField) {
    fmt.Printf("INFO: %s\n", msg)
}

func (l *MyLogger) Warn(msg string, fields ...openbymadata.LogField) {
    fmt.Printf("WARN: %s\n", msg)
}

func (l *MyLogger) Error(msg string, fields ...openbymadata.LogField) {
    fmt.Printf("ERROR: %s\n", msg)
}

func main() {
    // Create client with debug logger
    client := openbymadata.NewClient(&openbymadata.ClientOptions{
        Timeout: 15 * time.Second,
        Logger:  &MyLogger{},
    })

    // Make API calls - debug output will show automatically if DEBUG=true
    ctx := context.Background()
    indices, err := client.GetIndices(ctx)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Got %d indices\n", len(indices))
    }
}
```

## API Endpoints Covered

All API methods include debug logging:

- **Market Data**: `IsWorkingDay`, `GetIndices`, `MarketResume`
- **Securities**: `GetBluechips`, `GetGalpones`, `GetCedears`
- **Fixed Income**: `GetBonds`, `GetShortTermBonds`, `GetCorporateBonds`
- **Derivatives**: `GetOptions`, `GetFutures`
- **News**: `GetNews`, `GetIncomeStatement`

## Field Mapping Verification

The debug output shows both the raw API field names and how they're mapped to Go struct fields:

### Example: News API
```
Raw API fields: [descarga, emisor, fecha, referencia, simbolo, tipoArchivo]
Mapped to:
- emisor → Titulo (company name as title)
- referencia → Descripcion (reference as description)
- fecha → Fecha (date)
- descarga → Descarga (download URL)
```

## Detecting API Changes

If BYMA changes their API structure, debug mode will help you identify:
- New fields that appeared
- Fields that were renamed
- Fields that were removed
- Changed data types

## Demo

Run the debug demo to see it in action:

```bash
# Set DEBUG=true in .env first
go run cmd/debug/main.go
```

## Security Note

⚠️ **Warning**: Debug mode logs sensitive API response data. Only enable in development/testing environments, never in production. 