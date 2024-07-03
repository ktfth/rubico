package routes

import (
	"auth/models"
	"auth/utils"
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ServiceWeaver/weaver"
	"github.com/gin-contrib/sessions/cookie"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// AuthRoutes handles authentication routes.
type AuthRoutes struct {
	weaver.Implements[weaver.Main] // Altere para weaver.Main
	DB                             *mongo.Client
}

// RegisterLogin handles user registration/login.
// @Summary Register or login a user
// @Description Registers a new user or logs in an existing user with a magic link.
// @Tags auth
// @Accept json
// @Produce json
// @Param email body string true "User email address"
// @Param password body string true "User password"
// @Success 201 {object} map[string]string "Magic link sent successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /registerlogin [post]
func (r *AuthRoutes) RegisterLogin(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	// 1. Decodificar a requisição (JSON esperado)
	var user models.User
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		http.Error(w, "Erro ao decodificar a requisição", http.StatusBadRequest)
		return
	}

	// 2. Validar os dados do usuário (email, senha, etc.)
	if !utils.IsValidEmail(user.Email) {
		http.Error(w, "Email inválido", http.StatusBadRequest)
		return
	}

	// 3. Verificar se o usuário já existe no banco de dados
	filter := bson.M{"email": user.Email}
	var existingUser models.User
	err := r.DB.Database("rubico").Collection("users").FindOne(ctx, filter).Decode(&existingUser)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 3.1. Se o usuário não existir, criar um novo usuário
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Erro ao gerar hash de senha", http.StatusInternalServerError)
				return
			}
			user.Password = string(hash)
			user.ID = primitive.NewObjectID()
			_, err = r.DB.Database("rubico").Collection("users").InsertOne(ctx, user)
			if err != nil {
				http.Error(w, "Erro ao criar usuário no banco de dados", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Erro ao verificar usuário no banco de dados", http.StatusInternalServerError)
			return
		}
	} else {
		// 3.2. Se o usuário existir, verificar a senha
		if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
			return
		}
		// Atualiza o ID do usuário para o ID do usuário existente
		user.ID = existingUser.ID
	}

	// 4. Gerar um token único e seguro (com biblioteca apropriada)
	token, err := utils.GenerateToken(user.ID.Hex())
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	// 5. Salvar o token no banco de dados (com data de expiração)
	newToken := models.Token{
		ID:        primitive.NewObjectID(),
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Expira em 24 horas (ajuste conforme necessário)
		UserID:    user.ID,                        // Linkar o token ao usuário
	}
	_, err = r.DB.Database("rubico").Collection("tokens").InsertOne(ctx, newToken)
	if err != nil {
		http.Error(w, "Erro aosalvar token no banco de dados", http.StatusInternalServerError)
		return
	}

	// 6. Enviar o link mágico por e-mail (com biblioteca de email)
	if err := utils.SendMagicLinkEmail(user.Email, token); err != nil {
		http.Error(w, "Erro ao enviar o e-mail", http.StatusInternalServerError)
		return
	}

	// 7. Responder com sucesso
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Link mágico enviado com sucesso"})
}

// Verify handles the verification process when a user clicks on the magic link.
// @Summary      Verify magic link
// @Description  Verifies the magic link token and authenticates the user.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token   query      string  true  "Magic link token"
// @Success      200     {object}  map[string]string "User authenticated successfully"
// @Failure      400     {object}  map[string]string "Bad Request"
// @Failure      401     {object}  map[string]string "Unauthorized"
// @Failure      500     {object}  map[string]string "Internal server error"
// @Router       /verify [get]
func (r *AuthRoutes) Verify(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}

	var key = []byte(os.Getenv("SESSION_SECRET"))
	var store = cookie.NewStore(key)

	// 1. Obter o token da query string
	token := req.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token não fornecido", http.StatusBadRequest)
		return
	}
	log.Println("Verificando token:", token)

	// 2. Buscar o token no banco de dados
	filter := bson.M{"token": token}
	var tokenDoc models.Token
	if err := r.DB.Database("rubico").Collection("tokens").FindOne(ctx, filter).Decode(&tokenDoc); err != nil {
		http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
		return
	}

	// 3. Verificar se o token expirou
	if tokenDoc.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Token expirado", http.StatusUnauthorized)
		return
	}

	// 4. Obter o usuário associado ao token
	userFilter := bson.M{"_id": tokenDoc.UserID}
	var user models.User
	if err := r.DB.Database("rubico").Collection("users").FindOne(ctx, userFilter).Decode(&user); err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusInternalServerError)
		return
	}

	// 5. Autenticar o usuário (criar sessão, definir cookie, etc.)
	session, _ := store.Get(req, "auth-session") // Corrigindo a obtenção da sessão
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID.Hex()
	if err := session.Save(req, w); err != nil { // Salvar a sessão
		log.Println(err.Error())
		http.Error(w, "Erro ao salvar a sessão", http.StatusInternalServerError)
		return
	}

	// 6. Redirecionar o usuário para o app solicitante (opcional)
	redirectURL := req.URL.Query().Get("redirect_uri") // Obter a URL de redirecionamento da query string
	if redirectURL != "" {
		http.Redirect(w, req, redirectURL, http.StatusFound)
		return
	}

	// 7. Responder com sucesso (ou redirecionar)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuário autenticado com sucesso"})
}

// ValidateToken validates the provided token.
// @Summary      Validate token
// @Description  Validates the provided authentication token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token"
// @Success      200  {object}  map[string]interface{} "Token is valid"
// @Failure      401  {object}  map[string]string      "Unauthorized"
// @Router       /validatetoken [get]
func (r *AuthRoutes) ValidateToken(ctx context.Context, w http.ResponseWriter, req *http.Request) {
	// 1. Obter o token do cabeçalho Authorization
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Token não fornecido", http.StatusUnauthorized)
		return
	}

	// 2. Extrair o token do cabeçalho (geralmente no formato "Bearer <token>")
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		http.Error(w, "Formato de token inválido", http.StatusUnauthorized)
		return
	}
	token := tokenParts[1]

	// 3. Buscar o token no banco de dados
	filter := bson.M{"token": token}
	var tokenDoc models.Token
	if err := r.DB.Database("rubico").Collection("tokens").FindOne(ctx, filter).Decode(&tokenDoc); err != nil {
		http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
		return
	}

	// 4. Verificar se o token expirou
	if tokenDoc.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Token expirado", http.StatusUnauthorized)
		return
	}

	// 5. Token válido! (Você pode adicionar lógica adicional aqui, se necessário)
	log.Println("Token válido para o usuário:", tokenDoc.UserID)

	// 6. Responder com sucesso (ou informações do usuário, se necessário)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Token válido", "userID": tokenDoc.UserID})
}
