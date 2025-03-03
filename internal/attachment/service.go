package attachment

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	bucket "geolocation/pkg/s3"
	"geolocation/validation"
	"io/ioutil"
	"mime/multipart"
	"path"
	"strings"
)

type ServiceInterface interface {
	CreateAttachService(ctx context.Context, data *multipart.Form) error
	DeleteLogicAttachService(ctx context.Context, idStr string) error
}

type Service struct {
	repo           InterfaceRepository
	bucketFirebase string
}

func NewAttachmentService(repo InterfaceRepository, bucketFirebase string) *Service {
	return &Service{
		repo:           repo,
		bucketFirebase: bucketFirebase,
	}
}

func (s *Service) CreateAttachService(ctx context.Context, data *multipart.Form) error {
	var myForm AttachRequestCreate
	err := MapFormToStruct(data.Value, &myForm)
	if err != nil {
		return err
	}

	for _, files := range data.File {
		for _, fileHeader := range files {
			idFile := GetUUID()
			originalFilename := fileHeader.Filename
			fileExtension := strings.ToLower(path.Ext(originalFilename))
			newNameFileUp := idFile + fileExtension

			f, err := fileHeader.Open()
			if err != nil {
				return err
			}
			fileBytes, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return err
			}

			contentType := fileHeader.Header.Get("Content-Type")

			strUrl, err := bucket.UploadFileToS3(fileBytes, newNameFileUp, s.bucketFirebase, contentType)
			if err != nil {
				return err
			}

			userId, err := validation.ParseStringToInt64(myForm.UserId)
			if err != nil {
				return err
			}

			_, err = s.repo.CreateAttachments(ctx, db.CreateAttachmentsParams{
				UserID: userId,
				Description: sql.NullString{
					String: myForm.Description,
					Valid:  true,
				},
				Url: strUrl,
				NameFile: sql.NullString{
					String: originalFilename,
					Valid:  true,
				},
				SizeFile: sql.NullInt64{
					Int64: fileHeader.Size,
					Valid: true,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) DeleteLogicAttachService(ctx context.Context, idStr string) error {
	id, err := validation.ParseStringToInt64(idStr)
	if err != nil {
		return err
	}

	_, err = s.repo.GetAttachmentById(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("driver not found")
	}
	if err != nil {
		return err
	}

	err = s.repo.UpdateAttachmentLogicDelete(ctx, id)
	if err != nil {
		return err
	}

	//err = bucket.DeleteFile(ctx, s.bucketFirebase, infoAttach.NameFile.String)
	//if err != nil {
	//	return err
	//}

	return err
}
