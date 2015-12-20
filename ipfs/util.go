package ipfs

import (
    "fmt"
    "os"
)

func LogInfo(m string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "Info: %s\n", fmt.Sprintf(m, args...))
}

func LogWarn(m string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "Warn: %s\n", fmt.Sprintf(m, args...))
}

func LogError(m string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(m, args...))
}

func StripHash(hash string) string {
    if len(hash) < 7 { return hash }
    if hash[:6] == "/ipfs/" || hash[:6] == "/ipns/" { return hash[6:] }
    return hash
}

