package cloudinary

import (
	"context"
	"fmt"

	"github.com/backium/backend/core"
	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
)

type api struct {
	cld *cloudinary.Cloudinary
}

func New(url string) (core.Uploader, error) {
	c, err := cloudinary.NewFromURL(url)
	if err != nil {
		return nil, err
	}
	return &api{c}, nil
}

func (u *api) Upload(ctx context.Context, file string) (string, error) {
	res, err := u.cld.Upload.Upload(ctx, file, uploader.UploadParams{})
	if err != nil {
		return "", err
	}
	fmt.Printf("%+v", *res)
	return res.URL, nil
}
