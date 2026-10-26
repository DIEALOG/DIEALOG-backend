package usecase

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/DIEALOG/DIEALOG-backend/domain"
	"github.com/DIEALOG/DIEALOG-backend/ent"
	"github.com/google/uuid"
)

type profileUseCase struct {
	userRepository       domain.UserRepository
	userConfigRepository domain.UserConfigRepository
	contextTimeout       time.Duration
}

const maxProfileImageSize = 5 * 1024 * 1024

func NewProfileUseCase(userRepository domain.UserRepository, userConfigRepository domain.UserConfigRepository, timeout time.Duration) domain.ProfileUseCase {
	return &profileUseCase{
		userRepository:       userRepository,
		userConfigRepository: userConfigRepository,
		contextTimeout:       timeout,
	}
}

func (pu *profileUseCase) GetProfileByID(c context.Context, userID *domain.ID) (*domain.Profile, error) {
	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	user, err := pu.userRepository.FindByID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.Profile{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		PhoneNumber:       user.PhoneNumber,
		ProfileImagePath:  user.ProfileImagePath,
		Stakeholders:      user.Stakeholders,
		StakeholderGroups: user.StakeholderGroups,
	}, nil
}

func (pu *profileUseCase) UpdateProfileImage(c context.Context, userID *domain.ID, file *multipart.FileHeader) (*domain.Profile, error) {
	if file.Size <= 0 {
		return nil, domain.NewBadRequestError(errors.New("empty image file"))
	}
	if file.Size > maxProfileImageSize {
		return nil, domain.NewBadRequestError(errors.New("image size must be 5MB or less"))
	}

	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	src, err := file.Open()
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}
	defer src.Close()

	img, format, err := image.DecodeConfig(src)
	if err != nil {
		return nil, domain.NewBadRequestError(err)
	}

	_ = img

	extensions := map[string]string{
		"jpeg": ".jpg",
		"png":  ".png",
		"gif":  ".gif",
	}

	ext, ok := extensions[format]
	if !ok {
		return nil, domain.NewBadRequestError(errors.New("not support"))
	}

	// 이미지 파일을 처음부터 다시 읽기 위해 파일 시작 위치로 이동
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	dir := filepath.Join("public", "profiles")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	filename := uuid.NewString() + ext
	path := filepath.Join(dir, filename)

	dst, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return nil, domain.NewInternalServerError(err)
	}

	// 제한된 크기만큼 복사
	written, copyErr := io.Copy(
		dst,
		io.LimitReader(src, maxProfileImageSize+1),
	)
	closeErr := dst.Close()

	if written > maxProfileImageSize {
		_ = os.Remove(path)
		return nil, domain.NewBadRequestError(errors.New("image size must be 5MB or less"))
	}
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(path)
		return nil, domain.NewInternalServerError(errors.Join(copyErr, closeErr))
	}

	imagePath := fmt.Sprintf("/public/profiles/%s", filename)

	user, err := pu.userRepository.FindByID(ctx, userID)
	if err != nil {
		_ = os.Remove(path)
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	user.ProfileImagePath = &imagePath
	user, err = pu.userRepository.Update(ctx, user)
	if err != nil {
		_ = os.Remove(path)
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return &domain.Profile{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		PhoneNumber:       user.PhoneNumber,
		ProfileImagePath:  user.ProfileImagePath,
		Stakeholders:      user.Stakeholders,
		StakeholderGroups: user.StakeholderGroups,
	}, nil
}

func (pu *profileUseCase) UpdateUserConfig(c context.Context, config *domain.UserConfig, userID *domain.ID) (*domain.UserConfig, error) {
	ctx, cancel := context.WithTimeout(c, pu.contextTimeout)
	defer cancel()

	if config.UserID != *userID {
		return nil, domain.NewBadRequestError(errors.New("user id not match"))
	}

	_, err := pu.userConfigRepository.FindByUserID(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	cf, err := pu.userConfigRepository.Update(ctx, config)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, domain.NewNotFoundError(err)
		}
		return nil, domain.NewInternalServerError(err)
	}

	return cf, nil
}
