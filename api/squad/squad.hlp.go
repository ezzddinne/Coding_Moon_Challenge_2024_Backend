package squad

import (
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/ezzddinne/CMC_BACKEND/config"
)

// Image upload
func ImageUploadHelper(input interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(config.EnvCloudName(), config.EnvCloudAPIKey(), config.EnvCloudAPISecret())
	if err != nil {
		return "", err
	}

	currentTime := time.Now()

	//upload file
	uploadParam, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{AllowedFormats: []string{"jpg", "png"}, PublicID: "TeamLogo : " + currentTime.Format("2006-01-02 15:04:05"), Folder: config.EnvCloudUploadFolder()})
	if err != nil {
		return "", err
	}
	return uploadParam.SecureURL, nil
}

// File Upload
func FileUploadHelper(input interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//create cloudinary instance
	cld, err := cloudinary.NewFromParams(config.EnvCloudName(), config.EnvCloudAPIKey(), config.EnvCloudAPISecret())
	if err != nil {
		return "", err
	}

	currentTime := time.Now()

	//upload file
	uploadParam, err := cld.Upload.Upload(ctx, input, uploader.UploadParams{AllowedFormats: []string{"pdf"}, PublicID: "CV : " + currentTime.Format("2006-01-02 15:04:05"), Folder: config.EnvCloudUploadFolder()})
	if err != nil {
		return "", err
	}
	return uploadParam.SecureURL, nil
}
