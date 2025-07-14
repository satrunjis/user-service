package service

import (
	"context"
	"github.com/satrunjis/user-service/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
	"time"
)

var validSortBy = map[string]bool{"login": true, "reg_date": true}

const cost = 10

const (
	msgInvalidCharacters   = "contains invalid characters (allowed: a-z, A-Z, 0-9, _, -)"
	validCharactersPattern = `^[a-zA-Z0-9_-]+$`
)

func (s *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	normalizeUserFields(user)

	if err := prepareUserForCreation(user); err != nil {
		return err
	}

	if user.RegDate == nil {
		now := time.Now().UTC()
		user.RegDate = &now
	}

	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return mapRepositoryError(err, "create")
	}

	return nil
}

func (s *UserService) GetUserByID(ctx context.Context, id *string) (*domain.User, error) {
	if err := validationID(id); err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, mapRepositoryError(err, "get")
	}

	user.Password = nil

	return user, nil
}

func (s *UserService) SearchUsers(ctx context.Context, filters *domain.UserFilter) ([]*domain.User, error) {
	if filters != nil && filters.Size != nil && (*filters.Size <= 0 || *filters.Size > 100) {
		*filters.Size = 50
	}
	
	users, err := s.userRepo.Search(ctx, filters)
	if err != nil {
		return nil, mapRepositoryError(err, "search")
	}

	for i := range users {
		users[i].Password = nil
	}
	return users, nil
}

func (s *UserService) Replace(ctx context.Context, user *domain.User) error {
	if user.ID == nil {
		return NewServiceError(ErrCodeInvalidInput, "User ID is required")
	}

	normalizeUserFields(user)

	if err := prepareUserForCreation(user); err != nil {
		return err
	}

	err := s.userRepo.Replace(ctx, user)
	if err != nil {
		return mapRepositoryError(err, "replace")
	}

	return nil
}

func (s *UserService) UpdatePartial(ctx context.Context, user *domain.User) error {
	if user.RegDate != nil {
		return NewServiceError(ErrCodeInvalidInput, "Cannot update protected fields (registration date)")
	}

	normalizeUserFields(user)

	if err := prepareUserForCreation(user); err != nil {
		return err
	}

	err := s.userRepo.UpdatePartial(ctx, user)
	if err != nil {
		return mapRepositoryError(err, "update")
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id *string) error {
	if err := validationID(id); err != nil {
		return err
	}

	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		return mapRepositoryError(err, "delete")
	}

	return nil
}

func normalizeUserFields(user *domain.User) {
	if user.ID != nil && *user.ID == "" {
		user.ID = nil
	}
	if user.Login != nil && *user.Login == "" {
		user.Login = nil
	}
	if user.Username != nil && *user.Username == "" {
		user.Username = nil
	}
	if user.Password != nil && *user.Password == "" {
		user.Password = nil
	}
	if user.Description != nil && *user.Description == "" {
		user.Description = nil
	}
	if user.Comment != nil && *user.Comment == "" {
		user.Comment = nil
	}
	if user.SocialNet != nil && *user.SocialNet == "" {
		user.SocialNet = nil
	}
}

func hashPassword(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), cost)
	if err != nil {
		return "", NewServiceError(ErrCodeInternal, "Failed to hash password")
	}
	return string(bytes), nil
}

func prepareUserForCreation(user *domain.User) error {
	if err := validateUser(user); err != nil {
		return err
	}

	if user.Password != nil && *user.Password != "" {
		hashedPwd, err := hashPassword(*user.Password)
		if err != nil {
			return err
		}
		user.Password = &hashedPwd
	}
	return nil
}

func validationID(id *string) error {
	if id == nil || *id == "" {
		return NewServiceError(ErrCodeInvalidInput, "ID is required")
	}

	if len(*id) > 36 {
		return NewServiceError(ErrCodeInvalidInput, "ID must be less than 36 characters")
	}

	if !regexp.MustCompile(validCharactersPattern).MatchString(*id) {
		return NewServiceError(ErrCodeInvalidInput, "ID "+msgInvalidCharacters)
	}

	return nil
}

func validateUser(u *domain.User) error {
	var errs []string

	if u.ID != nil {
		if len(*u.ID) > 36 {
			errs = append(errs, "ID must be less than 36 characters")
		}
		if !regexp.MustCompile(validCharactersPattern).MatchString(*u.ID) {
			errs = append(errs, "ID "+msgInvalidCharacters)
		}
	}

	if u.Login != nil {
		login := *u.Login
		if len(login) < 5 || len(login) > 20 {
			errs = append(errs, "login must be 5-20 characters")
		}
		if !regexp.MustCompile(validCharactersPattern).MatchString(login) {
			errs = append(errs, "login "+msgInvalidCharacters)
		}
	}

	if u.Username != nil {
		username := *u.Username
		if len(username) > 50 {
			errs = append(errs, "username exceeds 50 character limit")
		}
	}

	if u.Password != nil {
		pwd := *u.Password
		if len(pwd) < 8 || len(pwd) > 64 {
			errs = append(errs, "password must be 8-64 characters")
		}
		if !regexp.MustCompile(validCharactersPattern).MatchString(pwd) {
			errs = append(errs, "password "+msgInvalidCharacters)
		}
	}

	if u.Description != nil && len(*u.Description) > 500 {
		errs = append(errs, "description exceeds 500 character limit")
	}

	if u.Comment != nil && len(*u.Comment) > 300 {
		errs = append(errs, "comment exceeds 300 character limit")
	}

	if u.RegDate != nil && u.RegDate.After(time.Now()) {
		errs = append(errs, "registration date cannot be in the future")
	}

	if u.Location != nil {
		loc := u.Location
		if loc.Lat < -90 || loc.Lat > 90 {
			errs = append(errs, "latitude must be between -90 and 90")
		}
		if loc.Lon < -180 || loc.Lon > 180 {
			errs = append(errs, "longitude must be between -180 and 180")
		}
	}

	if u.SocialNet != nil {
		validSocials := map[string]bool{
			"facebook":  true,
			"twitter":   true,
			"instagram": true,
			"max":       true,
			"vk":        true,
			"telegram":  true,
		}
		if !validSocials[strings.ToLower(*u.SocialNet)] {
			errs = append(errs, "invalid social network specified")
		}
	}

	if len(errs) != 0 {
		return NewServiceError(ErrCodeInvalidInput, strings.Join(errs, "; "))
	}
	return nil
}
