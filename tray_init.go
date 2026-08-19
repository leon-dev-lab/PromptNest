package main

func (a *App) startTray() {
	tray := NewTrayManager(a)
	a.tray = tray
	go tray.Run()
}
