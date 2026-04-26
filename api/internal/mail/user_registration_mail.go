package mail

type UserRegistrationMail struct {
	To             string
	URL            string
	Lang           string
	ExpiresMinutes int
}
