package foundation

import "golang.org/x/crypto/bcrypt"

type Hash struct{}

func NewHashable() Hash {
	return Hash{}
}
func (Hash) Make(value string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(value), 14)
	if err != nil {
		panic(err.Error())
	}

	return string(bytes)
}

func (Hash) Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
