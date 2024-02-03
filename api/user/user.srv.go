package user

import "github.com/go-playground/validator/v10"

var (
	validate = validator.New()
)

type mediaUpload interface {
	FileUpload(file File) (string, error)
}

type media struct{}

func NewMediaUpload() mediaUpload {
	return &media{}
}

func (*media) FileUpload(file File) (string, error) {
	//validate
	err := validate.Struct(file)
	if err != nil {
		return "", err
	}

	//upload
	uploadUrl, err := FileUploadHelper(file.File)
	if err != nil {
		return "", err
	}
	return uploadUrl, nil
}
