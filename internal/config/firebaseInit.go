package config

import (
	"context"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func InitFirebase() (*firebase.App, error) {
	ctx := context.Background()
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "firebase.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	return app, nil
}
