# Librería OpenBYMAData para Go

[![Go Reference](https://pkg.go.dev/badge/github.com/carvalab/openbymadata.svg)](https://pkg.go.dev/github.com/carvalab/openbymadata)
[![CI](https://github.com/carvalab/openbymadata/actions/workflows/go.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/go.yml)
[![Live API](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml)
[![Coverage](https://codecov.io/gh/carvalab/openbymadata/branch/main/graph/badge.svg)](https://app.codecov.io/gh/carvalab/openbymadata)
[![Release](https://badgen.net/github/release/carvalab/openbymadata)](https://github.com/carvalab/openbymadata/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🇬🇧 **English:** Go client for the BYMA (Buenos Aires stock exchange) free market data API. [Read the English README →](README.en.md)

Una librería completa en Go para acceder a datos financieros de la Bolsa de Comercio de Buenos Aires (BYMA - Bolsas y Mercados Argentinos) a través de su API pública gratuita.

Te da acceso a datos del mercado financiero argentino con tipado fuerte, concurrencia segura, caché incorporado y optimizaciones de rendimiento.

## Características

### 🚀 **Rendimiento y Diseño**
- **Caché Inteligente de 5 Minutos**: Las llamadas repetidas se responden desde memoria en lugar de volver a pegarle a la API
- **Búsquedas Individuales de Tickers**: Obtené valores específicos sin necesidad de recuperar colecciones completas
- **Operaciones por Lotes**: Recuperá múltiples valores de forma eficiente en una sola operación
- **Tipado Fuerte**: No más `interface{}` genéricos - estructuras adecuadas para cada instrumento financiero
- **Seguro para Concurrencia**: Seguro para usar en múltiples goroutines con caché thread-safe
- **Compatible con Context**: Todos los métodos aceptan `context.Context` para cancelación y timeouts
- **Lógica de Reintentos**: Retroceso exponencial incorporado para llamadas resilientes a la API
- **Manejo Completo de Errores**: Tipos de errores personalizados con lógica de reintentos

### 📊 **Cobertura de Datos del Mercado**
- **Acciones**: Líderes (blue chips), Panel general (galpones), CEDEARs, ETFs
- **Renta Fija**: Bonos gubernamentales y corporativos, letras, bonos verdes/sociales/sustentables, segmento SENEBI, índices de bonos, cronogramas de pago
- **Derivados**: Contratos de opciones, futuros
- **Datos Históricos**: Series temporales con OHLCV (Open, High, Low, Close, Volume) para gráficos
- **Datos de Mercado**: Índices y su historia (Merval), dólar MEP, mayores subas y bajas, búsqueda de símbolos, resumen del mercado, estado de días hábiles
- **Noticias y Financieros**: Noticias del mercado, estados de resultados

> **Nota**: "Securities" es un término financiero genérico en nuestro código Go, pero los endpoints reales de la API de BYMA son: `leading-equity`, `general-equity`, y `cedears`

### 🧪 **Pruebas y Confiabilidad**
- **Suite de Pruebas Completa**: Servidores HTTP de prueba para testing confiable
- **Benchmarks**: Incluye pruebas de rendimiento
- **Ejemplos**: Documentación re completa con ejemplos ejecutables

## Instalación

```bash
go get github.com/carvalab/openbymadata
```

## Inicio Rápido

📖 **La documentación completa con ejemplos está disponible directamente en el código y en [pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata)**

### Instalación

```bash
go get github.com/carvalab/openbymadata
```

### Ejemplo Básico

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/carvalab/openbymadata"
)

func main() {
    // Crear cliente (caché de 5 minutos habilitado por defecto)
    client := openbymadata.NewClient()
    ctx := context.Background()

    // Conseguir acción específica de EEUU (CEDEAR)
    aapl, err := client.GetCedear(ctx, "AAPL")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("🍎 AAPL: $%.2f (%.2f%%)\n", aapl.Last, aapl.Change)

    // Búsqueda universal (recomendada)
    security, err := client.GetSecurity(ctx, "BMA")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("🔍 BMA: $%.2f\n", security.Last)
}
```

### Funcionalidades Principales

**🎯 Búsquedas Individuales:** Obtené valores específicos sin descargar colecciones completas
```go
// Búsqueda universal (funciona para cualquier tipo de valor)
security, err := client.GetSecurity(ctx, "AAPL")

// Tipos específicos
aapl, err := client.GetCedear(ctx, "AAPL")     // CEDEARs (acciones de EEUU)
ggal, err := client.GetBluechip(ctx, "GGAL")  // Acciones líderes argentinas
```

**⚡ Operaciones por Lotes:** Múltiples valores en una sola operación
```go
watchlist := []string{"AAPL", "MSFT", "GOOGL", "GGAL"}
securities, err := client.GetMultipleSecurities(ctx, watchlist)
```

**📈 Datos Históricos:** Series temporales para gráficos
```go
// Últimos 30 días (diario)
historyData, err := client.GetHistoryLastDays(ctx, "SPY 24HS", 30)

// Rango personalizado con resolución tipada
weeklyData, err := client.GetHistory(ctx, "AAPL 24HS", openbymadata.ResolutionWeekly, from, to)

// Barras intradía (1 minuto por defecto, también 5 / 60)
bars, err := client.GetIntradayHistory(ctx, "GGAL 24HS", openbymadata.Resolution1Min, from, to)
```

**💲 Cotización directa y horario del mercado:**
```go
quote, err := client.GetCurrentQuote(ctx, "GGAL", openbymadata.Settlement24HS)
mt, err := client.GetMarketTime(ctx)
fmt.Println(mt.OpeningText, mt.ClosingText, mt.Timezone)
```

**💾 Caché Inteligente:** las llamadas repetidas dentro de 5 minutos se responden desde memoria

### Acceso Tradicional Basado en Colecciones

```go
// Conseguir todas las acciones líderes (en caché por 5 minutos)
bluechips, err := client.GetBluechips(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Encontradas %d acciones líderes:\n", len(bluechips))
for _, security := range bluechips[:5] { // Mostrar las primeras 5
    fmt.Printf("  %s: $%.2f (%.2f%%)\n", 
        security.Symbol, security.Last, security.Change)
}
```

### Configuración Personalizada

```go
// Crear cliente con opciones personalizadas
opts := &openbymadata.ClientOptions{
    Timeout:       30 * time.Second,
    RetryAttempts: 5,
    Logger:        customLogger, // Tu implementación de logger
}

client := openbymadata.NewClient(opts)
```

## Ejecutando Ejemplos

La librería incluye ejemplos completos que demuestran todas las funcionalidades:

### Correr la Demo Completa

```bash
# Clonar el repositorio
git clone https://github.com/carvalab/openbymadata.git
cd openbymadata

# Correr el ejemplo completo (muestra todas las funcionalidades)
go run cmd/example/main.go

# O compilar y ejecutar
go build -o byma-demo cmd/example/main.go
./byma-demo
```

### Ejemplo de Salida

```
🏛️  OpenBYMAData Go Library - Complete Example
============================================================
🚀 Features: Individual ticker lookups, 5-minute caching, batch operations

📊 1. Market Status & Info
------------------------------
Market Status: 🟢 OPEN
Market Indices (15):
  📈 G: 94518124.24 (0.01%)
  📈 M: 2213570.17 (0.01%)
  📈 SPBYCAP: 39874.07 (0.03%)

💰 2. Individual Ticker Lookups
-----------------------------------
🇺🇸 CEDEARs:
   🍎 AAPL: $13825.00 (0.00%) [220ms]
      Volume: 46576 | Last Update: 16:59:00
🇦🇷 Argentine Leading Equity:
   🏦 GGAL: $6620.00 (-0.00%)
🔍 Universal Search:
   📉 BMA: $9140.00 (-0.01%)
   📈 TSLA: $28175.00 (0.04%)
   ❌ UNKNOWN: Not found

📦 3. Batch Operations
-------------------------
💼 Portfolio (6/6 securities) [74µs]:
   🟢 AAPL  : $13825.00    +0.00%
   🔴 MSFT  : $22100.00    -0.00%
   🟢 GOOGL : $4300.00     +0.00%
   🟢 TSLA  : $28175.00    +0.04%
   🟢 META  : $38750.00    +0.01%
   🔴 GGAL  : $6620.00     -0.00%
   💰 Total Portfolio Value: $113770.00

📋 4. Collection Data (API endpoints: leading-equity, general-equity, cedears)
--------------------------------------------------------------------------------
💎 Leading Equity (21 securities from 'leading-equity' endpoint):
   🔴 ALUA: $709.00 (-0.00%) | Vol: 1171941
   ... and 18 more

🌎 CEDEARs (1132 securities from 'cedears' endpoint):
   🟢 AAL: $7530.00 (0.01%)
   ... and 1129 more

🏢 General Equity (178 securities from 'general-equity' endpoint):
   🟢 A3: $2500.00 (0.01%)
   ... and 175 more

🏛️  5. Fixed Income & Derivatives
-----------------------------------
📊 Government Bonds: 156 instruments
   Example: AL30 - $428.50
📈 Options: 2847 contracts
🔮 Futures: 23 contracts

⚡ 6. Cache Performance (5-minute automatic caching)
-------------------------------------------------------
🗄️  Cache Status:
   bluechips   : 21 items, age 215ms, fresh: true
   cedears     : 1132 items, age 256ms, fresh: true
   galpones    : 178 items, age 165ms, fresh: true

🏃 Cache Speed Test:
   Getting AAPL again (cached)... 750ns (lightning fast!)

📈 7. Historical Data (Chart Data)
-----------------------------------
📊 Historical Data for SPY (last 30 days):
   Retrieved 21 data points:
   First (2024-02-15): Open=$484.21 High=$486.58 Low=$483.12 Close=$485.22 Vol=45782
   Middle (2024-02-28): Open=$502.18 High=$503.47 Low=$501.25 Close=$502.87 Vol=52341
   Latest (2024-03-15): Open=$518.45 High=$519.23 Low=$517.89 Close=$518.67 Vol=38945

📅 Custom Date Range (Weekly data - last 3 months):
   AAPL Weekly Data - 13 weeks retrieved
   Latest week (2024-03-15): Close=$182.31

📰 8. News & Financial Data
------------------------------
📰 Latest News (24 items):
   📄 BYMA informa cotizaciones del día
      Date: 2024-03-15 18:30
   📄 Resultados trimestrales empresas listadas
      Date: 2024-03-15 16:45
📊 Income statements for ALUA: 8 records

🎉 Example Complete!
============================================================
✨ Features Demonstrated:
   • Individual ticker lookups (GetCedear, GetBluechip, GetSecurity)
   • Efficient batch operations (GetMultipleSecurities)
   • Historical data & charting (GetHistory, GetHistoryLastDays)
   • 5-minute automatic caching (fewer API calls)
   • API endpoint mapping:
     - GetBluechips()  → 'leading-equity' endpoint
     - GetGalpones()   → 'general-equity' endpoint
     - GetCedears()    → 'cedears' endpoint
     - GetHistory()    → 'chart/historical-series/history' endpoint
   • Full market data coverage (equities, bonds, derivatives)
   • Real-time market news and financial data
   • Thread-safe concurrent operations
   • Comprehensive error handling

🚀 Production Ready:
   • Context-aware operations
   • Built-in retry logic
   • Strongly-typed data structures
   • Zero external dependencies
```

### Correr Pruebas de Ejemplo

```bash
# Correr pruebas de ejemplo
go test -v -run "Example"

# Correr prueba de ejemplo específica
go test -v -run "ExampleClient"
```

## 📚 Documentación

### Recursos Disponibles

| Recurso | Descripción |
|---------|-------------|
| 📖 **[pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata)** | **Documentación principal** - Referencia completa de la API con ejemplos |
| 🎯 **[example_test.go](example_test.go)** | Ejemplos ejecutables para todas las funcionalidades |
| 🏛️ **[cmd/example/main.go](cmd/example/main.go)** | Demo completa mostrando toda la funcionalidad |
| 💾 **[CACHING_GUIDE.md](CACHING_GUIDE.md)** | Guía detallada de rendimiento del caché |
| 🐛 **[DEBUG.md](DEBUG.md)** | Guía de debugging y resolución de problemas |

### Visualizando Documentación

```bash
# 🌐 Mejor opción: Visitar pkg.go.dev (ejemplos y markdown)
# https://pkg.go.dev/github.com/carvalab/openbymadata

# Documentación en terminal
go doc -all

# Correr ejemplos
go test -v -run "Example"
```

## Referencia de la API

### Búsquedas Individuales de Tickers (¡NUEVO! 🔥)

```go
// Búsqueda universal (recomendada - busca en todos los tipos de valores)
security, err := client.GetSecurity(ctx, "AAPL")    // Funciona para cualquier símbolo

// Tipos específicos de valores
aapl, err := client.GetCedear(ctx, "AAPL")          // Acciones de EEUU (CEDEARs)
ggal, err := client.GetBluechip(ctx, "GGAL")        // Acciones líderes argentinas
galpone, err := client.GetGalpone(ctx, "SYMBOL")    // Panel general
bond, err := client.GetBond(ctx, "AL30")            // Todos los tipos de bonos
option, err := client.GetOption(ctx, "GGAL123")     // Opciones
future, err := client.GetFuture(ctx, "DOE25")       // Futuros
```

### Operaciones por Lotes (¡Eficiente! ⚡)

```go
// Conseguir múltiples valores eficientemente (usa caché compartida)
watchlist := []string{"AAPL", "MSFT", "GOOGL", "GGAL"}
securities, err := client.GetMultipleSecurities(ctx, watchlist)

// Buscar valores por símbolo parcial
results, err := client.SearchSecurities(ctx, "APP")  // Encuentra símbolos que contienen "APP"
```

### Datos Históricos y Gráficos (¡NUEVO! 📈)

```go
// Obtener datos históricos para los últimos 30 días (automáticamente agrega "24HS")
historyData, err := client.GetHistoryLastDays(ctx, "SPY 24HS", 30)

// Obtener datos históricos con rango de fechas personalizado
// Símbolos se normalizan automáticamente (se agrega "24HS" si no está presente)
// Resolución: "D" = diario, "W" = semanal, "M" = mensual
// from/to son time.Time (conversión a Unix automática)
from := time.Now().AddDate(0, -3, 0)  // 3 meses atrás
to := time.Now()                      // ahora
weeklyData, err := client.GetHistory(ctx, "AAPL 24HS", "W", from, to)

// Los datos retornan OHLCV como slices separados (formato más eficiente)
for i := range len(historyData.Time) {
    fmt.Printf("%s: Close=$%.2f, Volume=%d\n", 
        historyData.Time[i].Format("2006-01-02"), historyData.Close[i], historyData.Volume[i])
}
```

### Estado e Información del Mercado

```go
// Chequear si el mercado está operando hoy
isWorking, err := client.IsWorkingDay(ctx)

// Conseguir índices del mercado (Merval, etc.)
indices, err := client.GetIndices(ctx)

// Conseguir resumen del mercado
summary, err := client.MarketResume(ctx)

// Dólar MEP
mep, err := client.GetMEPIndices(ctx)
fmt.Printf("%s %.2f\n", mep[0].Symbol, mep[0].Last)

// Historia diaria del Merval ("M"; ver GetIndices para otros símbolos)
merval, err := client.GetIndexHistory(ctx, "M", openbymadata.ResolutionDaily, from, to)

// Mayores subas y bajas: acciones, bonos y opciones
movers, err := client.GetTopMovers(ctx)

// Buscar cualquier instrumento por símbolo o nombre (hasta 10 resultados)
matches, err := client.SearchInstruments(ctx, "galicia")
```

### Acceso Basado en Colecciones (Endpoints de la API)

```go
// Todos los valores de un tipo específico (en caché por 5 minutos)
bluechips, err := client.GetBluechips(ctx)  // → endpoint 'leading-equity'
galpones, err := client.GetGalpones(ctx)    // → endpoint 'general-equity'  
cedears, err := client.GetCedears(ctx)      // → endpoint 'cedears'
etfs, err := client.GetETFs(ctx)            // → endpoint 'etf'
```

### Renta Fija

```go
// Bonos gubernamentales
bonds, err := client.GetBonds(ctx)

// Letras de corto plazo
shortTermBonds, err := client.GetShortTermBonds(ctx)

// Bonos corporativos
corporateBonds, err := client.GetCorporateBonds(ctx)

// Bonos temáticos
green, err := client.GetGreenBonds(ctx)
social, err := client.GetSocialBonds(ctx)
sustainable, err := client.GetSustainabilityBonds(ctx)
linked, err := client.GetLinkedBonds(ctx)

// SENEBI: segmento bilateral sin contraparte central (símbolos terminan en ".SB")
senebi, err := client.GetSenebiBonds(ctx)
senebiLetters, err := client.GetSenebiLetters(ctx)
senebiCorp, err := client.GetSenebiCorporateBonds(ctx)

// Índices de renta fija y cronograma de pagos de un bono
bondIndices, err := client.GetBondIndices(ctx)
coupons, err := client.GetBondCoupons(ctx, "GD30")

// Los bonos ahora traen Currency, Market y DaysToMaturity
fmt.Println(bonds[0].Currency, bonds[0].Market, bonds[0].DaysToMaturity)
```

### Derivados

```go
// Contratos de opciones
options, err := client.GetOptions(ctx)

// Contratos de futuros
futures, err := client.GetFutures(ctx)
```

### Noticias y Datos Financieros

```go
// Noticias del mercado (en caché por 5 minutos)
news, err := client.GetNews(ctx)

// Estados de resultados para un ticker específico (en caché por símbolo)
statements, err := client.GetIncomeStatement(ctx, "GGAL")
```

### Gestión de Caché (¡NUEVO! 💾)

```go
// Conseguir información de caché
cacheInfo := client.GetCacheInfo()
fmt.Printf("Estado del caché: %+v\n", cacheInfo)

// Limpiar todos los datos en caché (fuerza llamadas frescas a la API)
client.ClearCache()

// Deshabilitar caché (no recomendado)
client := openbymadata.NewClient(&openbymadata.ClientOptions{
    EnableCache: false,
})
```

## Modelos de Datos

### Security
```go
type Security struct {
    Symbol         string    `json:"symbol"`
    Settlement     string    `json:"settlement"`
    BidSize        int64     `json:"bid_size"`
    Bid            float64   `json:"bid"`
    Ask            float64   `json:"ask"`
    AskSize        int64     `json:"ask_size"`
    Last           float64   `json:"last"`
    Close          float64   `json:"close"`
    Change         float64   `json:"change"`
    Open           float64   `json:"open"`
    High           float64   `json:"high"`
    Low            float64   `json:"low"`
    PreviousClose  float64   `json:"previous_close"`
    Turnover       float64   `json:"turnover"`
    Volume         int64     `json:"volume"`
    Operations     int64     `json:"operations"`
    DateTime       time.Time `json:"datetime"`
    Group          string    `json:"group"`
}
```

### Bond
```go
type Bond struct {
    // All Security fields plus:
    Expiration     time.Time `json:"expiration"`
}
```

### Option
```go
type Option struct {
    Symbol          string    `json:"symbol"`
    // ... price fields ...
    UnderlyingAsset string    `json:"underlying_asset"`
    Expiration      time.Time `json:"expiration"`
}
```

## Testing

### Correr Tests

```bash
# Correr todos los tests
go test ./...

# Correr tests con cobertura
go test -cover ./...

# Correr benchmarks
go test -bench=. ./...

# Correr contra la API real de BYMA
INTEGRATION_TEST=true go test -tags integration -run Integration ./...
```

### Estado de la API en Vivo

El [workflow de Integration](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml) corre la suite contra la API real de BYMA todos los días hábiles con el mercado abierto, así que el badge **Live API** de arriba te dice si la API de BYMA sigue siendo compatible con esta librería.

### Estructura de Tests

La librería utiliza servidores HTTP de prueba para simular respuestas de la API:

```go
func TestMyBusinessLogic(t *testing.T) {
    // Create test server with mock responses
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mockResponse := []openbymadata.Security{
            {Symbol: "GGAL", Last: 150.50},
        }
        json.NewEncoder(w).Encode(mockResponse)
    }))
    defer server.Close()

    // Create client pointing to test server
    client := openbymadata.NewClient(&openbymadata.ClientOptions{
        BaseURL: server.URL,
    })

    // Test your business logic
    result, err := myBusinessLogic(client)
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
}
```

## Manejo de Errores

La librería proporciona un manejo completo de errores con tipos de errores personalizados:

```go
securities, err := client.GetBluechips(ctx)
if err != nil {
    if bymaErr, ok := err.(*openbymadata.BYMAError); ok {
        switch bymaErr.Code {
        case "TIMEOUT":
            // Handle timeout
        case "RATE_LIMITED":
            // Handle rate limiting
        case "API_UNAVAILABLE":
            // Handle API unavailability
        default:
            log.Printf("API error: %v", bymaErr)
        }
    } else {
        log.Printf("Unexpected error: %v", err)
    }
}
```

## Rendimiento y Caché

### 🚀 Caché Inteligente de 5 Minutos

- **Caché Automático**: Todos los datos se almacenan en caché por 5 minutos por defecto
- **Llamadas Repetidas Rápidas**: Las llamadas en caché tardan microsegundos en lugar de un viaje por la red
- **Menos Llamadas a la API**: Menos carga sobre BYMA y menos riesgo de rate limiting
- **Thread-Safe**: Seguro para acceso concurrente a través de múltiples goroutines
- **Datos Frescos Garantizados**: El caché expira automáticamente después de 5 minutos

### Beneficios de Rendimiento

- **Búsquedas Individuales**: Obtené tickers específicos sin recuperar colecciones completas
- **Operaciones por Lotes**: Recuperá eficientemente múltiples valores usando caché compartido
- **Pooling de Conexiones**: Reutilización automática de conexiones HTTP
- **Lógica de Reintentos**: Retroceso exponencial incorporado para solicitudes fallidas
- **Soporte de Context**: Manejo adecuado de cancelación y timeouts

### Caché en Acción

```go
// First call - fetches from API (slow)
start := time.Now()
aapl, _ := client.GetCedear(ctx, "AAPL")
fmt.Printf("First call: %v\n", time.Since(start)) // ~100ms

// Second call - returns from cache (fast!)
start = time.Now()
aapl, _ = client.GetCedear(ctx, "AAPL")
fmt.Printf("Cached call: %v\n", time.Since(start)) // microsegundos, sin llamada de red

// Multiple securities use shared cache efficiently
securities, _ := client.GetMultipleSecurities(ctx, []string{"AAPL", "MSFT", "GOOGL"})
// All symbols returned instantly from cache!
```

### Acceso Concurrente

```go
// Example: Fetch multiple data types concurrently
var wg sync.WaitGroup
var bluechips []Security
var bonds []Bond
var indices []Index

wg.Add(3)

go func() {
    defer wg.Done()
    bluechips, _ = client.GetBluechips(ctx)    // Cached after first call
}()

go func() {
    defer wg.Done()
    bonds, _ = client.GetBonds(ctx)           // Cached after first call
}()

go func() {
    defer wg.Done()
    indices, _ = client.GetIndices(ctx)       // Cached after first call
}()

wg.Wait()
```

## Referencia a la Librería Python

Esta librería está inspirada en y proporciona funcionalidad equivalente a la librería original en Python [pyOBD](https://github.com/franco-lamas/PyOBD), con mejoras específicas de Go:

| Python pyOBD | Go openbymadata |
|---------------|-----------------|
| `pandas.DataFrame` | Strongly-typed structs |
| No type safety | Compile-time type checking |
| GIL limitations | True concurrency |
| No built-in retry | Exponential backoff retry |
| Basic error handling | Rich error types |
| Manual caching | Built-in 5-minute caching |

## Cómo Contribuir

Si BYMA cambió algo y la librería dejó de andar, [docs/MAINTAINING.md](docs/MAINTAINING.md) tiene los pasos para encontrarlo y arreglarlo.


1. Hacé un fork del repositorio
2. Creá tu rama de funcionalidad (`git checkout -b feature/funcionalidad-asombrosa`)
3. Sumá tests para tus cambios
4. Corré los tests: `go test ./...`
5. Commiteá tus cambios (`git commit -am 'Agrega funcionalidad asombrosa'`)
6. Pusheá a la rama (`git push origin feature/funcionalidad-asombrosa`)
7. Abrí un Pull Request

## Licencia

Este proyecto está licenciado bajo la Licencia MIT - mirá el archivo [LICENSE](LICENSE) para más detalles.

## Agradecimientos

- Librería original en Python [pyOBD](https://github.com/franco-lamas/PyOBD)
- [BYMA](https://www.byma.com.ar/) por proporcionar la API gratuita
- Comunidad de Go por excelentes herramientas y librerías

## Registro de Cambios

Las notas de cada versión están en [GitHub Releases](https://github.com/carvalab/openbymadata/releases).
