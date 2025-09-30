package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apache/answer-plugins/util"
	"github.com/apache/answer/plugin"
	"github.com/neermilov/apache-answer-telegram-connector/i18n"
	"github.com/segmentfault/pacman/log"
)

//go:embed info.yaml
var Info embed.FS

// Connector хранит конфигурацию Telegram-коннектора
type Connector struct {
	Config *ConnectorConfig
}

// ConnectorConfig содержит настройки плагина
type ConnectorConfig struct {
	BotToken     string `json:"bot_token"`
	BotUsername  string `json:"bot_username"`
	RedirectPath string `json:"redirect_path"`
}

// TelegramAuthData представляет данные, возвращаемые виджетом входа Telegram
type TelegramAuthData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// getEnvWithDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	config := &ConnectorConfig{
		BotToken:     getEnvWithDefault("TELEGRAM_BOT_TOKEN", ""),
		BotUsername:  getEnvWithDefault("TELEGRAM_BOT_USERNAME", ""),
		RedirectPath: getEnvWithDefault("TELEGRAM_REDIRECT_PATH", ""),
	}

	plugin.Register(&Connector{
		Config: config,
	})
}

// Info возвращает информацию о плагине
func (t *Connector) Info() plugin.Info {
	info := &util.Info{}
	info.GetInfo(Info)

	return plugin.Info{
		Name:        plugin.MakeTranslator(i18n.InfoName),
		SlugName:    info.SlugName,
		Description: plugin.MakeTranslator(i18n.InfoDescription),
		Author:      info.Author,
		Version:     info.Version,
		Link:        info.Link,
	}
}

// ConnectorLogoSVG возвращает SVG-лого Telegram
func (t *Connector) ConnectorLogoSVG() string {
	return `PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTkuNTM5OSA0LjgyQzE5LjE5OTkgNC44MiAxOC44Mjk5IDQuOTEgMTguNDk5OSA1LjA3TDQuOTU5ODkgMTAuOTNDMy45OTk4OSAxMS4zNiAzLjkzOTg5IDExLjczIDMuOTM5ODkgMTEuOTJDMy45Mzk4OSAxMi4xMiA0LjAyOTg5IDEyLjQ4IDQuOTU5ODkgMTIuOTJMMC4wODk4NDQgMTUuNjJDOC45OTg0IDE5LjcyIDEzLjE5OTkgMjEuODcgMTMuODk5OSAyMi4xN0MxNC42OTk5IDIyLjUyIDE1LjQ1OTkgMjIuNTEgMTUuOTg5OSAyMi4yMUMyMC41MDA5IDE5LjY3OTYgMjIuNzU2NiAxOC40MTMzIDIyLjc1OTkgMTguNDFDMjMuMDk5OSAxOC4yMSAyMy4zMzk5IDE3Ljc1IDIzLjIxOTkgMTcuMjVMMTkuODg5OSA1LjI0QzE5Ljc2OTkgNC43MiAxOS42ODk5IDQuODIgMTkuNTM5OSA0LjgyWk0xOS42NTk5IDcuMzRMMjIuMTI5OSAxNi4xOEwxMy4yMDk5IDE5LjkzTDEzLjkwOTkgMTUuNTRDMTMuOTM5OSAxNS4zNyAxNC4wMDk5IDE1LjIyIDE0LjEwOTkgMTUuMDlMMTkuNjU5OSA3LjM0Wk0xMi45OTk5IDEzLjYxTDEyLjM4OTkgMTcuNDFMMTAuNzU9OTkgMTYuNDQxMTEsIDEyLjQ2OTkgMTMuNjFIMS4xMjk5Wk0xMi4yMzku` // truncated long svg base64-like content
}

// ConnectorName возвращает название коннектора
func (t *Connector) ConnectorName() plugin.Translator {
	return plugin.MakeTranslator(i18n.ConnectorName)
}

// ConnectorSlugName возвращает slug коннектора
func (t *Connector) ConnectorSlugName() string {
	return "telegram"
}

// ConnectorSender генерирует HTML-страницу с виджетом входа Telegram
func (t *Connector) ConnectorSender(ctx *plugin.GinContext, receiverURL string) (redirectURL string) {
	// Вместо редиректа на Telegram рендерим HTML со встроенным стандартным виджетом.
	// После успешной аутентификации пользователь будет перенаправлен на receiverURL.

	botUsername := t.Config.BotUsername
	if botUsername == "" {
		log.Error("TELEGRAM_BOT_USERNAME not set, login widget won't work correctly")
	}

	htmlContent := fmt.Sprintf(`
		<div id="telegram-login-widget">
			<script async src="https://telegram.org/js/telegram-widget.js?22" 
				data-telegram-login="%s" 
				data-size="large" 
				data-radius="8"
				data-auth-url="%s" 
				data-request-access="write">
			</script>
		</div>
	`, botUsername, receiverURL)

	tmpl, err := template.New("telegram_login").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Telegram Login</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					margin: 0;
					padding: 0;
					display: flex;
					justify-content: center;
					align-items: center;
					min-height: 100vh;
					background-color: #f5f5f5;
				}
				.container {
					text-align: center;
					padding: 20px;
					background-color: white;
					border-radius: 8px;
					box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
				}
				h1 {
					margin-bottom: 20px;
					color: #333;
				}
				.message {
					margin-top: 20px;
					color: #666;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Telegram Login</h1>
				<div>{{.Widget}}</div>
				<div class="message">Click the button above to sign in with Telegram</div>
			</div>
		</body>
		</html>
	`)

	if err != nil {
		log.Error(err)
		return ""
	}

	// Отправляем HTML-ответ
	ctx.Writer.Header().Set("Content-Type", "text/html")
	ctx.Writer.WriteHeader(200)

	err = tmpl.Execute(ctx.Writer, map[string]interface{}{
		"Widget": template.HTML(htmlContent),
	})

	if err != nil {
		log.Error(err)
	}

	// Возвращаем пустую строку — вместо редиректа мы рендерим страницу
	return ""
}

// ConnectorReceiver обрабатывает данные, пришедшие от виджета Telegram
func (t *Connector) ConnectorReceiver(ctx *plugin.GinContext, receiverURL string) (userInfo plugin.ExternalLoginUserInfo, err error) {
	// Получаем параметры из query
	params := ctx.Request.URL.Query()

	id := params.Get("id")
	idInt, _ := strconv.ParseInt(id, 10, 64)

	authDate := params.Get("auth_date")
	authDateInt, _ := strconv.ParseInt(authDate, 10, 64)

	authData := TelegramAuthData{
		ID:        idInt,
		FirstName: params.Get("first_name"),
		LastName:  params.Get("last_name"),
		Username:  params.Get("username"),
		PhotoURL:  params.Get("photo_url"),
		AuthDate:  authDateInt,
		Hash:      params.Get("hash"),
	}

	// Проверяем подлинность данных
	if !t.verifyTelegramAuth(authData) {
		return userInfo, fmt.Errorf("telegram auth data verification failed")
	}

	// Проверка актуальности данных (не старше 1 дня)
	if time.Now().Unix()-authData.AuthDate > 86400 {
		return userInfo, fmt.Errorf("auth data is too old")
	}

	// Формируем информацию о пользователе из данных Telegram
	metaInfo, _ := json.Marshal(authData)
	userInfo = plugin.ExternalLoginUserInfo{
		ExternalID:  fmt.Sprintf("%d", authData.ID),
		DisplayName: strings.TrimSpace(fmt.Sprintf("%s %s", authData.FirstName, authData.LastName)),
		Username:    authData.Username,
		Email:       "", // Telegram не предоставляет email
		MetaInfo:    string(metaInfo),
		Avatar:      authData.PhotoURL,
	}

	// Если username пустой, используем ID как запасной вариант
	if userInfo.Username == "" {
		userInfo.Username = userInfo.ExternalID
	}

	return userInfo, nil
}

// verifyTelegramAuth проверяет подлинность данных от Telegram
func (t *Connector) verifyTelegramAuth(data TelegramAuthData) bool {
	if t.Config.BotToken == "" || data.Hash == "" {
		log.Warn("Missing bot token or hash")
		return false
	}

	// Собираем карту полей (без hash), чтобы сформировать проверочную строку
	dataMap := make(map[string]interface{})

	// Явно задаём строковые представления для некоторых полей
	dataMap["id"] = fmt.Sprintf("%d", data.ID)
	dataMap["auth_date"] = fmt.Sprintf("%d", data.AuthDate)
	if data.FirstName != "" {
		dataMap["first_name"] = data.FirstName
	}
	if data.LastName != "" {
		dataMap["last_name"] = data.LastName
	}
	if data.Username != "" {
		dataMap["username"] = data.Username
	}
	if data.PhotoURL != "" {
		dataMap["photo_url"] = data.PhotoURL
	}

	// Сортируем ключи по алфавиту
	keys := make([]string, 0, len(dataMap))
	for k := range dataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Формируем строку вида key=value\n...
	pairs := make([]string, 0, len(dataMap))
	for _, k := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, dataMap[k]))
	}
	dataCheckString := strings.Join(pairs, "\n")

	// Секретный ключ — SHA256 от токена бота
	secretKey := sha256.Sum256([]byte(t.Config.BotToken))

	// HMAC-SHA256 от проверочной строки
	h := hmac.New(sha256.New, secretKey[:])
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Сравниваем вычисленный хеш с полученным
	return calculatedHash == data.Hash
}
