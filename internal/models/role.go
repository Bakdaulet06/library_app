package models

type Role string

const (
	RoleClient   Role = "client"
	RoleAdmin    Role = "admin"
	RoleEmployee Role = "employee"
)
