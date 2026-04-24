package config

type Config interface {
	RegistrationTokenExpiresMinutes() int
}
