# Cache

Este paquete provee una interfaz sencilla para interactuar con Redis como sistema de cache en Go.

## Instalación

Agrega la dependencia en tu `go.mod`:

```
go get github.com/juanMaAV92/go-utils/cache
```

## Uso básico

```go
import (
    "github.com/juanMaAV92/go-utils/cache"
    "github.com/juanMaAV92/go-utils/log"
)

cfg := cache.CacheConfig{
    Host:       "localhost",
    Port:       "6379",
    ServerName: "my-service",
}
logger := log.NewLogger()
c, err := cache.New(cfg, logger)
if err != nil {
    panic(err)
}
ctx = context.Background()
```

## Opciones de Set y Get

Puedes modificar el comportamiento de las operaciones usando funciones de opciones:

```go
// Set con TTL y NX
c.Set(ctx, "key", "value", cache.WithTTL(time.Hour), cache.WithIfNotExist(true))

// Get y eliminar después de obtener
var cachedUser UserData
exists, err := c.Get(ctx, "key", cachedUser, cache.WithDeleteAfterGet(true))
```

## Configuración

- `CacheConfig`: estructura para configurar host, puerto y nombre del servicio.
- `SetOptions` y `GetOptions`: modifican el comportamiento de las operaciones.

## Licencia

MIT
