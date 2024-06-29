package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/resend/resend-go/v2"
	"os"
	"regexp"
)

// isValidEmail checks if the email address has a valid format.
func IsValidEmail(email string) bool {
	// Expressão regular para validar e-mail (básica)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// generateToken generates a random token for the magic link.
func GenerateToken(userID string) (string, error) {
	// Gera 32 bytes aleatórios para o token
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Codifica os bytes em base64 para criar o token
	token := base64.URLEncoding.EncodeToString(randomBytes)
	return token, nil
}

// sendMagicLinkEmail sends the magic link email using Resend.
func SendMagicLinkEmail(to, token string) error {
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev", // Endereço do remetente
		To:      []string{to},            // Endereço do destinatário
		Subject: "Seu link mágico de login",
		Html:    `Clique <a href="http://localhost:8080/verify?token=` + token + `">aqui</a> para fazer login.`,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
	return nil
}
