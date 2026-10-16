# openbymadata

[![Go Reference](https://pkg.go.dev/badge/github.com/carvalab/openbymadata.svg)](https://pkg.go.dev/github.com/carvalab/openbymadata)
[![CI](https://github.com/carvalab/openbymadata/actions/workflows/go.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/go.yml)
[![Live API](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml/badge.svg)](https://github.com/carvalab/openbymadata/actions/workflows/integration.yml)
[![Coverage](https://codecov.io/gh/carvalab/openbymadata/branch/main/graph/badge.svg)](https://app.codecov.io/gh/carvalab/openbymadata)
[![Release](https://badgen.net/github/release/carvalab/openbymadata)](https://github.com/carvalab/openbymadata/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🇬🇧 **English:** Go client for the BYMA (Buenos Aires stock exchange) free market data API. [Read the English README →](README.en.md)

Cliente en Go para [BYMA Open Data](https://open.bymadata.com.ar), la API
gratuita de datos de mercado de Bolsas y Mercados Argentinos. Acciones,
CEDEARs, ETFs, bonos, cauciones, opciones, futuros, índices, dólar MEP,
cotizaciones en vivo e históricos OHLCV, como structs tipados de Go. Sin API
key y sin dependencias.

El badge **Live API** es una suite de tests que corre contra la API real todos
los días hábiles, así te enterás si BYMA cambió algo antes que tus usuarios.

## Instalación

```bash
go get github.com/carvalab/openbymadata
```

Requiere Go 1.26 o superior.

## Uso rápido

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/carvalab/openbymadata"
)

func main() {
	client := openbymadata.NewClient() // caché de 5 minutos activado por defecto
	defer client.Close()
	ctx := context.Background()

	open, err := client.IsMarketOpen(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("mercado abierto:", open)

	ggal, err := client.GetBluechip(ctx, "GGAL")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GGAL %.2f (%+.2f%%)\n", ggal.Last, ggal.Change*100)

	// Los históricos necesitan el plazo: "GGAL 24HS", no "GGAL".
	bars, err := client.GetHistoryLastDays(ctx, "GGAL 24HS", 30)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("velas diarias:", len(bars.Time))

	mep, err := client.GetMEPIndices(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("dólar MEP: %.2f\n", mep[0].Last)
}
```

Hay ejemplos ejecutables de cada método en
[pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata#pkg-examples),
y `go run ./cmd/example` recorre casi toda la API.

## Lo que conviene saber

- **Los históricos necesitan el plazo en el símbolo.** `GetHistory`,
  `GetHistoryLastDays` y `GetIntradayHistory` reciben `"GGAL 24HS"`. Con
  `"GGAL"` solo devuelven un resultado vacío, sin error.
- **Para cotizaciones usá `Settlement24HS` o `SettlementCI`.** Con la
  liquidación T+1, `Settlement48HS` ya no trae precios.
- **Mercado abierto vs día hábil.** `IsMarketOpen` es true solo mientras BYMA
  está operando. `IsWorkingDay` dice si hoy es día hábil. `GetMarketTime` da el
  horario de apertura y cierre de hoy (UTC-3).
- **`Change` es una fracción.** `0.0123` significa +1,23%.
- **Los precios de cauciones y préstamos de valores son tasas** tal como las
  publica BYMA, y la escala cambia entre campos. Compará campos equivalentes.
- **El intradiario viene con demora** en la API gratuita: pocas velas por día.
- **`GetCompanyBalance` ya no anda** en la API gratuita; BYMA lo puso detrás de
  un login. Para los PDFs de los estados contables usá `GetIncomeStatement`.

## La API de un vistazo

Las listas de todo el mercado se guardan 5 minutos en caché (ver
[Caché](#caché)). Las consultas por símbolo, cotizaciones, históricos y estado
del mercado siempre van a la API.

**Estado del mercado e índices**

| Método | Devuelve |
|---|---|
| `IsMarketOpen(ctx)` | si BYMA está operando ahora |
| `IsWorkingDay(ctx)` | si hoy es día hábil |
| `GetMarketTime(ctx)` | horario de apertura y cierre de hoy |
| `GetServerTime(ctx)` | el reloj de BYMA (UTC-3) |
| `GetIndices(ctx)` | valores de índices (el Merval es `"M"`) |
| `GetIndexHistory(ctx, symbol, res, from, to)` | OHLCV de un índice |
| `GetMEPIndices(ctx)` | índices de dólar MEP |
| `MarketResume(ctx)` | totales negociados |
| `GetTopMovers(ctx)` | acciones, bonos y opciones del panel de subas y bajas |

**Renta variable**

| Método | Devuelve |
|---|---|
| `GetBluechips(ctx)` | panel líder |
| `GetGalpones(ctx)` | panel general |
| `GetCedears(ctx)` | CEDEARs |
| `GetETFs(ctx)` | ETFs |
| `GetCorporateGovernance(ctx)`, `GetCorporateGovernancePlus(ctx)` | paneles de gobierno corporativo |

**Renta fija y mercado de dinero**

| Método | Devuelve |
|---|---|
| `GetBonds(ctx)` | bonos públicos |
| `GetShortTermBonds(ctx)` | letras |
| `GetCorporateBonds(ctx)` | obligaciones negociables |
| `GetGreenBonds`, `GetSocialBonds`, `GetSustainabilityBonds`, `GetLinkedBonds` | bonos temáticos |
| `GetSenebiBonds`, `GetSenebiLetters`, `GetSenebiCorporateBonds` | segmento SENEBI (símbolos `.SB`) |
| `GetFinancialTrusts(ctx)` | fideicomisos financieros |
| `GetRepos(ctx)` | cauciones |
| `GetSecuritiesLending(ctx)` | préstamos de valores |
| `GetBondIndices(ctx)` | índices de renta fija |
| `GetBondCoupons(ctx, symbol)` | cronograma de renta y amortización de un bono |

**Derivados**

| Método | Devuelve |
|---|---|
| `GetOptions(ctx)` | opciones, con `OptionType` CALL/PUT |
| `GetFutures(ctx)` | futuros (suele venir vacío en la API gratuita) |

**Cotizaciones e históricos**

| Método | Devuelve |
|---|---|
| `GetCurrentQuote(ctx, symbol, settlement)` | punta compradora/vendedora y último precio |
| `GetHistory(ctx, symbol, res, from, to)` | OHLCV (diario, semanal, mensual o intradiario) |
| `GetHistoryLastDays(ctx, symbol, days)` | OHLCV diario hasta hoy |
| `GetIntradayHistory(ctx, symbol, res, from, to)` | velas de 1, 5 o 60 minutos |

**Búsquedas**

| Método | Devuelve |
|---|---|
| `GetSecurity(ctx, symbol)` | una acción o CEDEAR de cualquier panel |
| `GetBluechip`, `GetCedear`, `GetGalpone`, `GetBond`, `GetOption`, `GetFuture` | un instrumento por símbolo |
| `GetMultipleSecurities(ctx, symbols)` | varias acciones de una vez |
| `SearchSecurities(ctx, text)` | acciones cuyo símbolo contiene el texto |
| `SearchInstruments(ctx, query)` | cualquier instrumento por símbolo o nombre, hasta 10 |
| `GetInstrumentCatalog(ctx)` | todos los instrumentos listados (unos 11.000) |

**Empresas y presentaciones**

| Método | Devuelve |
|---|---|
| `GetCompanyInfo(ctx, symbol)` | datos de la sociedad |
| `GetEquityProfile(ctx, symbol)` | ficha de la especie (ISIN, valor nominal, ...) |
| `GetCompanyManagement(ctx, symbol)` | directorio |
| `GetMarketResponsibles(ctx, symbol)` | responsables de relaciones con el mercado |
| `GetIncomeStatement(ctx, symbol)` | estados contables con link al PDF |
| `GetNews(ctx)` | avisos de BYMA con link al PDF |

La documentación campo por campo de cada tipo está en
[pkg.go.dev](https://pkg.go.dev/github.com/carvalab/openbymadata).

## Configuración

```go
opts := openbymadata.DefaultClientOptions() // caché activo, timeout 30s, 3 reintentos
opts.Timeout = 10 * time.Second
opts.Logger = myLogger // cualquier cosa con Debug/Info/Warn/Error
client := openbymadata.NewClient(opts)
```

Si armás `ClientOptions` a mano, poné `EnableCache: true` para tener caché; el
valor cero lo apaga.

### Caché

Las listas de todo el mercado (paneles, bonos, opciones, índices, noticias,
MEP, el catálogo) se guardan en memoria 5 minutos y se comparten entre
goroutines. Las búsquedas como `GetBluechip` o `GetMultipleSecurities` leen de
esas listas.

```go
client.GetCacheInfo() // cantidad, antigüedad y vigencia por clave
client.ClearCache()   // forzar datos nuevos
```

Más en [CACHING_GUIDE.md](CACHING_GUIDE.md).

### Errores, reintentos y cancelación

- Todos los métodos reciben un `context.Context`. Si lo cancelás o vence su
  deadline, se corta el request en curso y cualquier reintento pendiente;
  chequealo con `errors.Is(err, context.DeadlineExceeded)`.
- Los errores de red y los HTTP 400, 408, 429 y 5xx se reintentan
  (`RetryAttempts`, 3 por defecto) esperando 1s, 2s, 3s… BYMA responde 400 al
  azar cuando está cargada, por eso se reintenta. El resto de los 4xx fallan
  enseguida.
- Los errores son errores de Go envueltos, por ejemplo
  `request failed after 2 attempts: HTTP error 404`.

Para ver cada request, pasá un `Logger`; ver [DEBUG.md](DEBUG.md).

## Tests

```bash
go test ./...                                                    # tests unitarios, respuestas grabadas de BYMA
INTEGRATION_TEST=true go test -tags integration -run Integration ./...  # API real
```

Los tests unitarios reproducen respuestas grabadas de la API real
(`internal/api/testdata`), así que corren sin conexión. La suite contra la API
real corre en CI todos los días hábiles; los chequeos de precios solo corren
si `IsMarketOpen` es true, así que feriados y corridas fuera de horario no
fallan.

## Seguirle el ritmo a BYMA

BYMA cambia la API sin avisar. Hay tres cosas que lo detectan:

- la corrida **Live API** de los días hábiles,
- un chequeo que falla cuando BYMA deja de mandar un campo que parseamos,
- una corrida semanal **Discover** que compara la web de BYMA con esta
  librería y abre un issue con los endpoints para agregar, sacar o corregir.

[docs/MAINTAINING.md](docs/MAINTAINING.md) tiene el paso a paso de cada caso.

## Cómo Contribuir

Issues y pull requests son bienvenidos. Corré `go test ./...` y `gofmt -s`
antes de abrir un PR. Si BYMA cambió algo, empezá por
[docs/MAINTAINING.md](docs/MAINTAINING.md).

## Agradecimientos

- [PyOBD](https://github.com/franco-lamas/PyOBD), la librería de Python de la
  que salió esta
- [BYMA](https://www.byma.com.ar/) por la API gratuita

## Licencia

MIT, ver [LICENSE](LICENSE). Las notas de cada versión están en
[GitHub Releases](https://github.com/carvalab/openbymadata/releases).
