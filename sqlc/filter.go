package db

import (
	"math"
)

const MaxFilterSize = 500

type Filter struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type UserActionFilter struct {
	Filter
	User *User
}

type UserListFilter struct {
	Filter
	Search struct {
		Email         string
		FirstName     string
		LastName      string
		BusinessName  string
		AccountType   string
		AccountNumber string
	}
}

type BusinessListFilter struct {
	Filter
	Search        string
	Status        string
	CreatedAtFrom string
	CreatedAtTo   string
}

var EmptyMetadata = Metadata{}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func (f Filter) Limit() int {
	if f.PageSize == 0 || f.PageSize < 0 || f.PageSize > MaxFilterSize {
		return MaxFilterSize
	}
	return f.PageSize
}

func (f Filter) Offset() int {
	if f.Page == 0 || f.Page < 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.PageSize
}

func (f UserListFilter) Limit() int {
	if f.PageSize == 0 {
		return 10
	} else if f.PageSize > MaxFilterSize {
		return MaxFilterSize
	}
	return f.PageSize
}

func (f UserListFilter) Offset() int {
	if f.Page == 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.PageSize
}

func (f BusinessListFilter) Limit() int {
	if f.PageSize == 0 {
		return 100
	} else if f.PageSize > MaxFilterSize {
		return MaxFilterSize
	}
	return f.PageSize
}

func (f BusinessListFilter) Offset() int {
	if f.Page == 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.PageSize
}

func CalculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
