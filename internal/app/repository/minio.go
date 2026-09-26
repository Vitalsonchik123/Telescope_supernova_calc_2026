package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// InitMinio — вызывается один раз при старте приложения
func (r *Repository) InitMinio(endpoint, accessKey, secretKey, bucket string) error {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}
	r.minio = client
	r.minioBucket = bucket

	// Создаём бакет, если его нет
	exists, err := client.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
		// Делаем публичным
		policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + bucket + `/*"]}]}`
		client.SetBucketPolicy(context.Background(), bucket, policy)
	}
	return nil
}

// UploadFile — загружает файл в MinIO и возвращает имя файла (ключ)
func (r *Repository) UploadFile(prefix string, id uint, header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("не открыть файл: %w", err)
	}
	defer file.Close()

	// Проверяем Content-Type по первым 512 байтам
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("не прочитать файл: %w", err)
	}
	contentType := http.DetectContentType(buffer)
	_, _ = file.Seek(0, 0)

	// Формируем имя файла
	filename := prefix + "_" + strconv.FormatUint(uint64(id), 10) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + extFromContentType(contentType)

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucket,
		filename,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки в minio: %w", err)
	}
	return filename, nil
}

// getMinioURL — публичный URL файла
func (r *Repository) GetMinioURL(filename string) string {
	if filename == "" {
		return ""
	}
	return "http://localhost:9000/" + r.minioBucket + "/" + filename
}

func extFromContentType(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ""
	}
}
