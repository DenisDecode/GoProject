// Точка входа для запуска командой `go run .` из корня проекта (требование этапа 1).
// Вся логика находится в internal/app, поэтому эта точка входа и cmd/server
// запускают одно и то же приложение.
package main

import "expense-tracker/internal/app"

func main() {
	app.Run()
}
