# Paquete pointers

Helpers para trabajar con punteros en Go.


## Funcionalidad
- Conversión entre valores y punteros de tipos básicos.
- Facilita el manejo seguro de punteros.

## Funciones disponibles
- `Pointer[T any](v T) *T`: Devuelve el puntero de cualquier valor.
- `FirstNonNil[T any](values ...*T) *T`: Retorna el primer puntero no nulo de la lista, o nil si todos son nil.

## Ejemplo de uso
```go
import "github.com/juanMaAV92/go-utils/pointers"

ptr := pointers.Int(42)
```
