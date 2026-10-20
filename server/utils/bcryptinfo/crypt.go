package bcryptinfo

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Crypt struct {
}

func NewCrypt() *Crypt {
	return &Crypt{}
}

func (c *Crypt) HashPassword(password string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}
	return string(hashPwd), err
}

func (c *Crypt) CheckPasswords(password, hashPwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashPwd), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
