//go:build windows

package main

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

//go:embed build/windows/icon.ico
var trayIconBytes []byte

const (
	wmDestroy       = 0x0002
	wmClose         = 0x0010
	wmNull          = 0x0000
	wmLButtonDblClk = 0x0203
	wmRButtonUp     = 0x0205
	wmApp           = 0x8000
	trayCallbackMsg = wmApp + 42

	nimAdd    = 0x00000000
	nimModify = 0x00000001
	nimDelete = 0x00000002

	nifMessage = 0x00000001
	nifIcon    = 0x00000002
	nifTip     = 0x00000004
	nifInfo    = 0x00000010

	imageIcon      = 1
	lrLoadFromFile = 0x00000010

	mfString    = 0x00000000
	mfSeparator = 0x00000800

	tpmRightButton = 0x0002
	tpmNonotify    = 0x0080
	tpmReturnCmd   = 0x0100

	niifInfo = 0x00000001
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	shell32              = syscall.NewLazyDLL("shell32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	pRegisterClassExW    = user32.NewProc("RegisterClassExW")
	pCreateWindowExW     = user32.NewProc("CreateWindowExW")
	pDefWindowProcW      = user32.NewProc("DefWindowProcW")
	pDestroyWindow       = user32.NewProc("DestroyWindow")
	pGetMessageW         = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessageW    = user32.NewProc("DispatchMessageW")
	pPostQuitMessage     = user32.NewProc("PostQuitMessage")
	pPostMessageW        = user32.NewProc("PostMessageW")
	pLoadImageW          = user32.NewProc("LoadImageW")
	pDestroyIcon         = user32.NewProc("DestroyIcon")
	pCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	pAppendMenuW         = user32.NewProc("AppendMenuW")
	pTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	pDestroyMenu         = user32.NewProc("DestroyMenu")
	pGetCursorPos        = user32.NewProc("GetCursorPos")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	pShellNotifyIconW    = shell32.NewProc("Shell_NotifyIconW")
	pGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
	activeTrayMu         sync.Mutex
	activeTray           *TrayManager
)

type point struct {
	X int32
	Y int32
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	Private uint32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UTimeout         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         guid
	HBalloonIcon     uintptr
}

type TrayManager struct {
	app      *App
	mu       sync.Mutex
	hwnd     uintptr
	hicon    uintptr
	nid      notifyIconData
	iconPath string
}

func NewTrayManager(app *App) *TrayManager { return &TrayManager{app: app} }

func (t *TrayManager) Run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	activeTrayMu.Lock()
	activeTray = t
	activeTrayMu.Unlock()
	defer func() {
		activeTrayMu.Lock()
		if activeTray == t {
			activeTray = nil
		}
		activeTrayMu.Unlock()
	}()

	className, _ := syscall.UTF16PtrFromString("PromptNestTrayWindow")
	windowName, _ := syscall.UTF16PtrFromString("PromptNest Tray")
	hinst, _, _ := pGetModuleHandleW.Call(0)
	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   syscall.NewCallback(trayWndProc),
		HInstance:     hinst,
		LpszClassName: className,
	}
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	hwnd, _, _ := pCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		0, 0, 0, 0, 0,
		0, 0, hinst, 0,
	)
	if hwnd == 0 {
		return
	}

	_ = os.MkdirAll(t.app.store.DataDirectory, 0o755)
	iconPath := filepath.Join(t.app.store.DataDirectory, "PromptNestTray.ico")
	_ = os.WriteFile(iconPath, trayIconBytes, 0o644)
	iconPtr, _ := syscall.UTF16PtrFromString(iconPath)
	hicon, _, _ := pLoadImageW.Call(0, uintptr(unsafe.Pointer(iconPtr)), imageIcon, 32, 32, lrLoadFromFile)

	nid := notifyIconData{
		CbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           nifMessage | nifIcon | nifTip,
		UCallbackMessage: trayCallbackMsg,
		HIcon:            hicon,
	}
	copyUTF16(nid.SzTip[:], "PromptNest")
	pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&nid)))

	t.mu.Lock()
	t.hwnd = hwnd
	t.hicon = hicon
	t.nid = nid
	t.iconPath = iconPath
	t.mu.Unlock()

	var m msg
	for {
		ret, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (t *TrayManager) Stop() {
	t.mu.Lock()
	hwnd := t.hwnd
	t.mu.Unlock()
	if hwnd != 0 {
		pPostMessageW.Call(hwnd, wmClose, 0, 0)
	}
}

func (t *TrayManager) Notify(title, body string) {
	t.mu.Lock()
	if t.hwnd == 0 {
		t.mu.Unlock()
		return
	}
	nid := t.nid
	t.mu.Unlock()
	nid.UFlags = nifInfo
	copyUTF16(nid.SzInfoTitle[:], title)
	copyUTF16(nid.SzInfo[:], body)
	nid.DwInfoFlags = niifInfo
	nid.UTimeout = 1800
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&nid)))
}

func (t *TrayManager) showMenu() {
	t.mu.Lock()
	hwnd := t.hwnd
	t.mu.Unlock()
	if hwnd == 0 {
		return
	}
	menu, _, _ := pCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer pDestroyMenu.Call(menu)
	appendMenu(menu, mfString, 1001, "显示 PromptNest")
	appendMenu(menu, mfString, 1002, "打开数据目录")
	pAppendMenuW.Call(menu, mfSeparator, 0, 0)
	appendMenu(menu, mfString, 1003, "退出")
	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(menu, tpmRightButton|tpmNonotify|tpmReturnCmd, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	pPostMessageW.Call(hwnd, wmNull, 0, 0)
	switch cmd {
	case 1001:
		go t.app.showWindow()
	case 1002:
		go func() { _ = t.app.openDataFolder() }()
	case 1003:
		go t.app.quitFromTray()
	}
}

func (t *TrayManager) cleanup() {
	t.mu.Lock()
	nid := t.nid
	hicon := t.hicon
	t.hwnd = 0
	t.mu.Unlock()
	if nid.HWnd != 0 {
		pShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))
	}
	if hicon != 0 {
		pDestroyIcon.Call(hicon)
	}
}

func trayWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	activeTrayMu.Lock()
	t := activeTray
	activeTrayMu.Unlock()
	if t != nil {
		switch message {
		case trayCallbackMsg:
			switch uint32(lParam) {
			case wmLButtonDblClk:
				go t.app.showWindow()
			case wmRButtonUp:
				t.showMenu()
			}
			return 0
		case wmClose:
			t.cleanup()
			pDestroyWindow.Call(hwnd)
			return 0
		case wmDestroy:
			pPostQuitMessage.Call(0)
			return 0
		}
	}
	ret, _, _ := pDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func appendMenu(menu uintptr, flags uintptr, id uintptr, text string) {
	ptr, _ := syscall.UTF16PtrFromString(text)
	pAppendMenuW.Call(menu, flags, id, uintptr(unsafe.Pointer(ptr)))
}

func copyUTF16(dst []uint16, text string) {
	u, _ := syscall.UTF16FromString(text)
	if len(u) > len(dst) {
		u = u[:len(dst)]
		u[len(u)-1] = 0
	}
	copy(dst, u)
}
