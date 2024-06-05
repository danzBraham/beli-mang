package controllers

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	media_entity "github.com/danzBraham/beli-mang/internal/entities/media"
	http_helper "github.com/danzBraham/beli-mang/internal/helpers/http"
	"github.com/danzBraham/beli-mang/internal/http/middlewares"
	"github.com/google/uuid"
)

type MediaController struct{}

func NewMediaController() *MediaController {
	return &MediaController{}
}

func (c *MediaController) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	isAdmin, ok := r.Context().Value(middlewares.ContextIsAdminKey).(bool)
	if !ok {
		http_helper.ResponseError(w, http.StatusUnauthorized, "IsAdmin type assertion failed", "IsAdmin not found in the context")
		return
	}
	if !isAdmin {
		http_helper.ResponseError(w, http.StatusUnauthorized, "Unauthorized error", "you're not admin")
		return
	}

	err := r.ParseMultipartForm(media_entity.MaxUploadSize)
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Unable to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, err.Error(), "Unable to get file from form")
		return
	}
	defer file.Close()

	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	if fileExt != ".jpg" && fileExt != ".jpeg" {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", "File must be in .jpg or .jpeg format")
		return
	}

	if header.Size < media_entity.MinUploadSize || header.Size > media_entity.MaxUploadSize {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", "File size must be between 10KB and 2MB")
		return
	}

	filename := uuid.New().String() + fileExt

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", err.Error())
		return
	}

	client := s3.NewFromConfig(cfg)

	uploader := manager.NewUploader(client)
	result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_S3_BUCKET_NAME")),
		Key:    aws.String(filename),
		Body:   file,
		ACL:    "public-read",
	})
	if err != nil {
		http_helper.ResponseError(w, http.StatusBadRequest, "Bad request error", err.Error())
		return
	}

	http_helper.ResponseSuccess(w, http.StatusOK, "File uploaded sucessfully", &media_entity.UploadImageResponse{
		ImageURL: result.Location,
	})
}
