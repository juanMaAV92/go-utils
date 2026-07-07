# go-utils

[![CI](https://github.com/juanMaAV92/go-utils/actions/workflows/ci.yml/badge.svg)](https://github.com/juanMaAV92/go-utils/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/juanMaAV92/go-utils?color=6ee7a8&label=release)](https://github.com/juanMaAV92/go-utils/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/juanMaAV92/go-utils.svg)](https://pkg.go.dev/github.com/juanMaAV92/go-utils)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-juanmaav92.github.io-6ee7a8)](https://juanmaav92.github.io/go-utils)

Librería de utilidades en Go para construir microservicios en AWS. Módulo único, patrones consistentes en todos los paquetes: `ConfigFromEnv`, diseño guiado por interfaces y trazabilidad distribuida con OpenTelemetry.

> Documentación interactiva: **https://juanmaav92.github.io/go-utils**
>
> Ejemplo real: [go-echo-blueprint](https://github.com/juanMaAV92/go-echo-blueprint) — template de microservicio Echo con arquitectura por capas que integra base de datos (PostgreSQL), caché (Redis), mensajería (SQS/SNS), tracing distribuido, validación y middleware de identidad.

```bash
go get github.com/juanMaAV92/go-utils/v2
```

```go
import "github.com/juanMaAV92/go-utils/v2/logger"
```

Requiere Go 1.25+.

---

## v2.0.0

El path del módulo ahora incluye el sufijo `/v2` (versionado semántico de importaciones en Go). Actualiza tus imports a `github.com/juanMaAV92/go-utils/v2/...`. Lo más destacado de esta versión:

- **logger** — los IDs de OTel `trace_id`/`span_id` ahora se inyectan correctamente en cada línea de log (se corregió un bug donde se omitían silenciosamente debido a la envoltura del handler).
- **telemetry** — el muestreo fraccional es `ParentBased`, por lo que se respetan las decisiones de muestreo tomadas aguas arriba (upstream) entre servicios.
- **security/jwt** — la validación ahora obliga a usar exclusivamente RS256, exige la presencia de `exp` y valida el emisor (`iss`).
- **httpclient** — se deshabilitó el logueo de los cuerpos de request/response para evitar la fuga accidental de credenciales y tokens.
- **cache/redis** — `WithKeepTTL` funciona correctamente (antes causaba un error de sintaxis garantizado); `AddToSet` ahora rechaza las opciones no soportadas.
- **messaging/scheduler** — `UpdateSchedule` ya no vuelve a habilitar schedules que estaban pausados; la zona horaria se aplica de forma correcta y se eliminó el campo no funcional `Tags` a favor de un flag `Disabled`.
- **messaging/sqs/consumer** — se implementó backoff de sondeo (poll) ante errores, un vaciado gradual (graceful drain) al apagar (evitando entregas duplicadas) y límites de AWS aplicados a `MaxMessages`/`WaitTimeSeconds`.
- **database/postgresql** — los errores ahora envuelven su causa original (permitiendo usar `errors.Is`); se añadió configuración independiente para `MaxIdleConns` y se aumentó el tamaño por defecto del pool de conexiones.
- **middleware/identity** — la propagación de permisos con scope específico ya no filtra el comodín general `all:all`.

---

## Paquetes

### Core

| Paquete | Descripción |
|---|---|
| [`env`](env/) | Parseo de variables de entorno con conversión de tipos y valores seguros por defecto |
| [`errors`](errors/) | Respuestas estructuradas de error HTTP; `errors/echo` para el handler de errores de Echo |
| [`logger`](logger/) | Logger estructurado basado en `log/slog` con inyección automática de trazas/spans de OTel |
| [`telemetry`](telemetry/) | Inicialización del SDK de OpenTelemetry (exportador OTLP, sampler y recurso base) |
| [`validator`](validator/) | Wrapper de `go-playground/validator` que retorna respuestas estructuradas de error |
| [`pointers`](pointers/) | Helpers genéricos para punteros (`Pointer[T]`, `Value[T]`, `FirstNonNil`) |
| [`timeutil`](timeutil/) | Utilidades para formateo de fechas y conversión a punteros |
| [`httpclient`](httpclient/) | Cliente HTTP (Resty) con propagación de trazas OTel y políticas de reintento |
| [`security/jwt`](security/jwt/) | Generación y validación de tokens JWT RS256 con claims genéricos |

### Infraestructura

| Paquete | Descripción |
|---|---|
| [`database/postgresql`](database/postgresql/) | Wrapper de GORM con operaciones CRUD, paginación, transacciones y tracing OTel |
| [`cache/redis`](cache/redis/) | Cliente Redis con TTL, operaciones de sets, Pub/Sub y métricas OTel |
| [`storage/s3`](storage/s3/) | Cliente S3: `GetObject`, `PutObject`, `DeleteObject`, `HeadObject` y generación de URLs firmadas |

### Mensajería

| Paquete | Descripción |
|---|---|
| [`messaging/sqs`](messaging/sqs/) | Cliente SQS: productor (envío simple/batch) y consumidor (worker pool con desempaquetado de SNS) |
| [`messaging/sns`](messaging/sns/) | Productor SNS con propagación de contexto de trazas W3C en atributos |
| [`messaging/scheduler`](messaging/scheduler/) | EventBridge Scheduler: ejecuciones únicas de Lambda, ventanas flexibles y política de reintento |

### Middleware y Testing

| Paquete | Descripción |
|---|---|
| [`middleware/identity`](middleware/identity/) | Middleware para Echo que propaga la identidad del usuario desde headers HTTP; helpers de RBAC |
| [`testutil/echo`](testutil/echo/) | Helpers para pruebas table-driven de handlers Echo (`PrepareContext`, `ToJSONString`) |
| [`testutil/http`](testutil/http/) | Helpers de pruebas HTTP agnósticos del framework (`AssertStatus`, `AssertJSONField`, `DecodeJSON`) |

---

## Diseño

- **Observabilidad nativa** — cada paquete emite spans de OpenTelemetry; el logger inyecta `trace_id`/`span_id` automáticamente desde el contexto.
- **ConfigFromEnv(prefix)** — patrón consistente en todos los paquetes de AWS; el prefijo aísla las variables de entorno por instancia del cliente.
- **Guiado por interfaces** — expone interfaces con implementaciones no exportadas; las interfaces facilitan mocks en pruebas unitarias sin interactuar con AWS real.
- **Sin credenciales fijas** — todos los paquetes de AWS usan la cadena de credenciales estándar (variables de entorno, archivo config, IAM roles).

---

## Licencia

MIT
