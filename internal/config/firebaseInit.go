package config

import (
	"context"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

const fileName = "firebase.json"

func InitFirebase() (*firebase.App, error) {
	if _, err := os.Stat(fileName); err != nil {
		return nil, err
	}

	opt := option.WithCredentialsFile(fileName)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	return app, nil
}
