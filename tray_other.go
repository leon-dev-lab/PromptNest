//go:build !windows

package main

type TrayManager struct{}

func NewTrayManager(app *App) *TrayManager       { return &TrayManager{} }
func (t *TrayManager) Run()                      {}
func (t *TrayManager) Stop()                     {}
func (t *TrayManager) Notify(title, body string) {}
