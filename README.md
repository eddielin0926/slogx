# slogx

A simple colorful `slog` handler.

## Installation

```bash
go get github.com/eddielin0926/slogx
```

## Usage

```go
id := 1
opts := &slogx.Options{Level: slog.LevelDebug}
logger := slog.New(slogx.New(os.Stdout, opts)).With("front", fmt.Sprintf(" [%d]", id))
logger.Debug("This is a debug message")
logger.Info("This is an info message")
logger.Warn("This is a warning message")
logger.Error("This is an error message")
logger.Info("This is an info message",
    "int", 1,
    "float", 1.1,
    "string", "hello",
    "person", struct {
        Name string
        Age  int
    }{"John", 30},
)
```
