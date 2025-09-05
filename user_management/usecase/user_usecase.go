package usecase

import (
	"context"
	"errors"
	"io"
	"time"
	"user_management/domain"
	"user_management/infrastructure/auth"
)

type UserUsecase struct {
	userRepo      domain.UserRepository
	imageUploader domain.ImageUploader
	hasher       *auth.PasswordHasher
}

func NewUserUsecase(ur domain.UserRepository, iu domain.ImageUploader) domain.UserUsecase {
	return &UserUsecase{
		userRepo:      ur,
		imageUploader: iu,
		hasher:       &auth.PasswordHasher{},
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

func (upd *UserUsecase) ProfileUpdate(ctx context.Context, userid string, gender string, birthDate time.Time, languagePreference string, file io.Reader) error {
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

func (upd *UserUsecase) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return upd.userRepo.FindByID(ctx, id)
}

func (upd *UserUsecase) ChangePassword(ctx context.Context, user_id string, oldPassword string, newPassword string) error {
	user, err := upd.userRepo.FindByID(ctx, user_id)
	if err != nil {
		return err
	}

	if !upd.hasher.CompareHashAndPassword(user.Password, oldPassword) {
		return errors.New("invalid credentials")
	}

	hashedPassword, err := upd.hasher.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return upd.userRepo.UpdateUserPassword(ctx, user.Email, hashedPassword)
}
