package hwnd

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modUser32 = windows.NewLazySystemDLL("User32.dll")

	procPostQuitMessage  = modUser32.NewProc("PostQuitMessage")
	procPostMessage      = modUser32.NewProc("PostMessageW")
	procCreateWindowEx   = modUser32.NewProc("CreateWindowExW")
	procDefWindowProc    = modUser32.NewProc("DefWindowProcW")
	procDispatchMessage  = modUser32.NewProc("DispatchMessageW")
	procGetMessage       = modUser32.NewProc("GetMessageW")
	procPeekMessage      = modUser32.NewProc("PeekMessageW")
	procTranslateMessage = modUser32.NewProc("TranslateMessage")
	procRegisterClassExW = modUser32.NewProc("RegisterClassExW")
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cnClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       syscall.Handle
}
type wndMsg struct {
	hwnd     syscall.Handle
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}
type point struct {
	x, y int32
}

func createNativeWindow(wndProc WndProc) (syscall.Handle, error) {
	className, err := syscall.UTF16PtrFromString("GoWindow")
	if err != nil {
		return 0, err
	}
	wcls := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(wndProc),
		lpszClassName: className,
	}
	cls, err := registerClassEx(&wcls)
	if err != nil {
		return 0, err
	}
	const WS_OVERLAPPED = 0x00000000
	hwnd, err := createWindowEx(0,
		cls,
		"Unknown",
		WS_OVERLAPPED,
		0, 0, 300, 200,
		0,
		0,
		0,
		0)
	if err != nil {
		return 0, err
	}
	return hwnd, nil
}

func createWindowEx(dwExStyle uint32, lpClassName uint16, lpWindowName string, dwStyle uint32, x, y, w, h int32, hWndParent, hMenu, hInstance syscall.Handle, lpParam uintptr) (syscall.Handle, error) {
	wname, err := syscall.UTF16PtrFromString(lpWindowName)
	if err != nil {
		return 0, err
	}
	hwnd, _, err := syscall.Syscall12(
		procCreateWindowEx.Addr(),
		12,
		uintptr(dwExStyle),
		uintptr(lpClassName),
		uintptr(unsafe.Pointer(wname)),
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(w),
		uintptr(h),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam))
	if hwnd == 0 {
		return 0, fmt.Errorf("CreateWindowEx failed: %w", err)
	}
	return syscall.Handle(hwnd), nil
}

func peekMessage(m *wndMsg, hwnd syscall.Handle, wMsgFilterMin, wMsgFilterMax, wRemoveMsg uint32) bool {
	r, _, _ := syscall.Syscall6(
		procPeekMessage.Addr(),
		5,
		uintptr(unsafe.Pointer(m)),
		uintptr(hwnd),
		uintptr(wMsgFilterMin),
		uintptr(wMsgFilterMax),
		uintptr(wRemoveMsg), 0)

	return r != 0
}

func getMessage(m *wndMsg, hwnd syscall.Handle, wMsgFilterMin, wMsgFilterMax uint32) (int32, error) {
	r, _, err := syscall.Syscall6(
		procGetMessage.Addr(),
		4,
		uintptr(unsafe.Pointer(m)),
		uintptr(hwnd),
		uintptr(wMsgFilterMin),
		uintptr(wMsgFilterMax), 0, 0)

	if int32(r) == -1 {
		return -1, err
	}

	return int32(r), nil
}

func translateMessage(m *wndMsg) {
	syscall.Syscall(procTranslateMessage.Addr(), 1, uintptr(unsafe.Pointer(m)), 0, 0)
}

func dispatchMessage(m *wndMsg) {
	syscall.Syscall(procDispatchMessage.Addr(), 1, uintptr(unsafe.Pointer(m)), 0, 0)
}

func registerClassEx(cls *wndClassEx) (uint16, error) {
	a, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(cls)))
	if a == 0 {
		return 0, err
	}
	return uint16(a), nil
}

func postQuitMessage(nExitCode int) (err error) {
	r1, _, e1 := syscall.Syscall(procPostQuitMessage.Addr(), 1, uintptr(nExitCode), 0, 0)
	if r1 == 0 {
		err = e1
	}
	return
}
func postMessage(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall6(procPostMessage.Addr(), 4, uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam), 0, 0)
	if r1 == 0 {
		err = e1
	}
	return
}

func DefWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	r, _, _ := syscall.Syscall6(procDefWindowProc.Addr(), 4, uintptr(hwnd), uintptr(msg), wparam, lparam, 0, 0)
	return r
}
