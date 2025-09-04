package usecase

import (
	"context"
	"errors"
	"user_management/domain"
	"io"
)

type UserUsecase struct {
	userRepo      domain.UserRepository
	imageUploader domain.ImageUploader
}

func NewUserUsecase(ur domain.UserRepository, iu domain.ImageUploader) domain.UserUsecase {
	return &UserUsecase{
		userRepo:      ur,
		imageUploader: iu,
	}
}

var (
	ErrSelfRoleChange = errors.New("cannot change own role")
	ErrAlreadyHasRole = errors.New("user already has the target role")
)

func (upd *UserUsecase) Promote(ctx context.Context, userid, Email string) error {
	targetUser, err := upd.userRepo.FindByEmail(ctx, Email)
	if err != nil {
		return err
	}

	if targetUser.ID == userid {
		return ErrSelfRoleChange
	}

	if targetUser.Role == string(domain.RoleAdmin) {
		return ErrAlreadyHasRole
	}

	return upd.userRepo.UpdateUserRole(ctx, string(domain.RoleAdmin), Email)
}

func (upd *UserUsecase) Demote(ctx context.Context, userid, Email string) error {
	targetUser, err := upd.userRepo.FindByEmail(ctx, Email)
	if err != nil {
		return err
	}

	if targetUser.ID == userid {
		return ErrSelfRoleChange
	}

	if targetUser.Role == string(domain.RoleUser) {
		return ErrAlreadyHasRole
	}

	return upd.userRepo.UpdateUserRole(ctx, string(domain.RoleUser), Email)
}

func (upd *UserUsecase) ProfileUpdate(ctx context.Context, userid string, gender string, birthDate string, languagePreference string, file io.Reader) error {
	user, err := upd.userRepo.FindByID(ctx, userid)
	if err != nil {
		return err
	}

	imageURL, err := upd.imageUploader.UploadImage(ctx, file, "profile")
	if err != nil {
		return errors.New("failed to upload image")
	}

	return upd.userRepo.UpdateUserProfile(ctx, gender, birthDate, languagePreference, imageURL, user.Email)
}

func (upd *UserUsecase) GetAllUsers(ctx context.Context, page int, limit int) ([]domain.User, int64, error) {
	return upd.userRepo.GetAllUsers(ctx, page, limit)
}
