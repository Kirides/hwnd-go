package hwnd

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
)

type WndProc func(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr

type Hwnd struct {
	Handle syscall.Handle
}

func (h Hwnd) ProcessMessagesContext(ctx context.Context) error {
	const WM_QUIT = 0x12
	const EXIT_OK = 0

	done, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-ctx.Done():
			cancel()
		case <-done.Done():
		}
		postMessage(h.Handle, WM_QUIT, EXIT_OK, 0)
	}()

	for {
		var msg wndMsg
		r, err := getMessage(&msg, h.Handle, 0, 0)
		if r == 0 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("GetMessage failed. %w", err)
		}
		// GetMessage handles WM_QUIT by returning r == 0
		// If one were to use PeekMessage, you'd have to handle it
		// on your own
		//     if msg.message == WM_QUIT {
		//         return nil
		//     }

		translateMessage(&msg)
		dispatchMessage(&msg)
	}
}

func New(wndProc WndProc) (Hwnd, error) {
	// Call win32 API from a single OS thread.
	runtime.LockOSThread()
	hwnd, err := createNativeWindow(wndProc)
	if err != nil {
		return Hwnd{}, err
	}
	return Hwnd{hwnd}, nil
}
