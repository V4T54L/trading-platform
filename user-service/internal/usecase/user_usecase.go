package usecase

import (
	"context"
	"errors"
	"gotrade/user-service/internal/domain"
	"gotrade/user-service/internal/repository"
	"gotrade/user-service/pkg/utils"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type UserUsecase interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUserProfile(ctx context.Context, userID int64) (*domain.User, error)
}

type userUsecase struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewUserUsecase(userRepo repository.UserRepository, jwtSecret string) UserUsecase {
	return &userUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (uc *userUsecase) Register(ctx context.Context, email, password string) (*domain.User, error) {
	existingUser, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    email,
		Password: hashedPassword,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *userUsecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token, err := utils.GenerateJWT(user.ID, uc.jwtSecret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (uc *userUsecase) GetUserProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

