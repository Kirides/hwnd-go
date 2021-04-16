## Introduction

lightweight windows HWND implementation for usage with Win32 APIs that require a hWnd.

E.g. [AddClipboardFormatListener (winuser.h)](https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-addclipboardformatlistener)

## Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/kirides/hwnd-go"
)

func main() {
    // Create a hWnd with a respective WndProc
    h, err := hwnd.New(func(h syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
        // Handle message

        // Print incoming message
        fmt.Printf("WndProc(%v, %s, %v, %v)\n", h, wm.String(msg), wParam, lParam)

        // Should always call hwnd.DefWindowProc if not handled otherwise
        // refer to general Win32 API programming
        return hwnd.DefWindowProc(h, msg, wParam, lParam)
    })
    if err != nil {
        panic(err)
    }

    // support graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
    defer cancel()

    // Process "window" message queue
    h.ProcessMessagesContext(ctx)
}
```