package main

import (
	"bytes"
	"text/template"
)

const (
	migrationTemplate = `package db

import (
	"database/sql"
	"time"

	"github.com/tcfw/go-migrate"
)

func init() {
	register(migrate.NewSimpleMigration(
		"{{.Name}}",
		time.Date({{.Year}}, {{.Month}}, {{.Day}}, {{.Hour}}, {{.Minute}}, {{.Second}}, 0, time.Local),

		//Up
		func(tx *sql.Tx) error {
			//Write your 'up' migration
			return nil
		},

		//Down
		func(tx *sql.Tx) error {
			//Write your 'down' migration
			return nil
		},
	))
}`
)

type migrationData struct {
	Name   string
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

func getTemplate(data *migrationData) (string, error) {
	tmpl, err := template.New("migrationTemplate").Parse(migrationTemplate)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)

	err = tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return string(buf.Bytes()), nil
}
