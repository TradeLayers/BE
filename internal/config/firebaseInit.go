package config

import (
	"context"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitFirebase() (*firebase.App, error) {
	const firebaseConfig = "firebase.json"
	if _, err := os.Stat(firebaseConfig); err != nil {
		return nil, err
	}

	opt := option.WithCredentialsFile(firebaseConfig)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	return app, nil
}
