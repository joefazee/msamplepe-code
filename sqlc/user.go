package db

import (
	"fmt"
	"strings"
)

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetPhone() string {
	return u.Phone
}

func (u *User) AccountName() string {

	if strings.ToLower(u.AccountType) == "business" || strings.ToLower(u.AccountType) == "affiliate_merchant" {
		return u.BusinessName
	}
	if u.MiddleName != "" {
		return fmt.Sprintf("%s %s %s", u.FirstName, u.LastName, u.MiddleName)
	}

	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}
