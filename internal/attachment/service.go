package attachment

import (
	"context"
	"database/sql"
	"errors"
	db "geolocation/db/sqlc"
	bucket "geolocation/pkg/s3"
	"io/ioutil"
	"mime/multipart"
	"path"
	"strings"
)

type ServiceInterface interface {
	CreateAttachService(ctx context.Context, data *multipart.Form, userID int64) error
	UpdateAttachService(ctx context.Context, userID int64, data *multipart.Form) error
	GetAllAttachmentById(ctx context.Context, userID int64, origin string) ([]Attachment, error)
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

func (s *Service) CreateAttachService(ctx context.Context, data *multipart.Form, userID int64) error {
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

			_, err = s.repo.CreateAttachments(ctx, db.CreateAttachmentsParams{
				UserID: userID,
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
				Type: myForm.Type,
			})
			if err != nil {
				return err
			}

			if myForm.Type == "user" {
				err := s.repo.UpdateProfilePictureByUserId(ctx, db.UpdateProfilePictureByUserIdParams{
					ProfilePicture: sql.NullString{
						String: strUrl,
						Valid:  true,
					},
					ID: userID,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Service) UpdateAttachService(ctx context.Context, userID int64, data *multipart.Form) error {
	var myForm AttachRequestCreate
	err := MapFormToStruct(data.Value, &myForm)
	if err != nil {
		return err
	}

	_, err = s.repo.GetAttachmentById(ctx, db.GetAttachmentByIdParams{
		UserID: userID,
		Type:   myForm.Type,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("attachment not found")
	}
	if err != nil {
		return err
	}

	err = s.repo.UpdateAttachmentLogicDelete(ctx, db.UpdateAttachmentLogicDeleteParams{
		UserID: userID,
		Type:   myForm.Type,
	})
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

			_, err = s.repo.CreateAttachments(ctx, db.CreateAttachmentsParams{
				UserID: userID,
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

			if myForm.Type == "user" {
				err := s.repo.UpdateProfilePictureByUserId(ctx, db.UpdateProfilePictureByUserIdParams{
					ProfilePicture: sql.NullString{
						String: strUrl,
						Valid:  true,
					},
					ID: userID,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) GetAllAttachmentById(ctx context.Context, userID int64, origin string) ([]Attachment, error) {
	results, err := s.repo.GetAllAttachmentById(ctx, db.GetAllAttachmentByIdParams{
		UserID: userID,
		Type:   origin,
	})
	if err != nil {
		return nil, err
	}

	var result []Attachment
	for _, attachment := range results {
		result = append(result, Attachment{
			UserID: attachment.UserID,
			URL:    attachment.Url,
		})
	}

	return result, nil
}
