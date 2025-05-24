package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config map .env configs to this struct
type Config struct {
	Timezone             string        `mapstructure:"TIMEZONE"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	HttServerAddress     string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFERESH_TOKEN_DURATION"`
	DBMigrationURL       string        `mapstructure:"MIGRATION_URL"`
	Environment          string        `mapstructure:"ENVIRONMENT"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	RedisUsername        string        `mapstructure:"REDIS_USERNAME"`
	RedisPassword        string        `mapstructure:"REDIS_PASSWORD"`
	RedisDB              int           `mapstructure:"REDIS_DB"`
	MailDomain           string        `mapstructure:"MAIL_DOMAIN"`
	MailTemplates        string        `mapstructure:"MAIL_TEMPLATES"`
	MailHost             string        `mapstructure:"MAIL_HOST"`
	MailPort             int           `mapstructure:"MAIL_PORT"`
	MailUsername         string        `mapstructure:"MAIL_USERNAME"`
	MailPassword         string        `mapstructure:"MAIL_PASSWORD"`
	MailEncryption       string        `mapstructure:"MAIL_ENCRYPTION"`
	MailFromAddress      string        `mapstructure:"MAIL_FROM_ADDRESS"`
	MailFromName         string        `mapstructure:"MAIL_FROM_NAME"`
	MailAPI              string        `mapstructure:"MAIL_API"`
	MailAPIKey           string        `mapstructure:"MAIL_API_KEY"`
	MailAPIUrl           string        `mapstructure:"MAIL_API_URL"`
	FlutterWaveAPIKey    string        `mapstructure:"FLUTTERWAVE_API_KEY"`
	PaystackSecretKey    string        `mapstructure:"PAYSTACK_SECRET_KEY"`
	WalletSymmetricKey   string        `mapstructure:"WALLET_SYMMETRIC_KEY"`
	AppBaseURL           string        `mapstructure:"APP_BASE_URL"`
	UploadPath           string        `mapstructure:"UPLOAD_PATH"`
	UploadBaseURL        string        `mapstructure:"UPLOAD_BASE_URL"`

	// AWS credentials
	AWSAccessKeyID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion          string `mapstructure:"AWS_REGION"`
	AWSEndpoint        string `mapstructure:"AWS_ENDPOINT"`
	FileBucket         string `mapstructure:"FILE_BUCKET"`

	MaxFileUploadSize  int64 `mapstructure:"MAX_FILE_UPLOAD_SIZE"`
	MaxBodyPayloadSize int64 `mapstructure:"MAX_BODY_PAYLOAD_SIZE"`

	AllowedMimeTypes   []string `mapstructure:"ALLOWED_MIME_TYPES"`
	FileUploadProvider string   `mapstructure:"FILE_UPLOAD_PROVIDER"`

	ExchangeRatePrecision int32 `mapstructure:"EXCHANGE_RATE_PRECISION"`

	// Polaris credentials
	PolarisSecretKey     string `mapstructure:"POLARIS_SECRET_KEY"`
	PolarisAPIKey        string `mapstructure:"POLARIS_API_KEY"`
	PolarisMockMode      string `mapstructure:"POLARIS_MOCK_MODE"`
	PolarisWebhookURL    string `mapstructure:"POLARIS_WEBHOOK_URL"`
	PolarisAccountNumber string `mapstructure:"POLARIS_ACCOUNT_NUMBER"`

	// Fidelity credentials
	FidelitySecretKey string `mapstructure:"FIDELITY_SECRET_KEY"`
	FidelityAPIKey    string `mapstructure:"FIDELITY_API_KEY"`
	FidelityMockMode  string `mapstructure:"FIDELITY_MOCK_MODE"`
	// FidelityWebhookURL    string `mapstructure:"FIDELITY_WEBHOOK_URL"`
	FidelityAccountNumber string `mapstructure:"FIDELITY_ACCOUNT_NUMBER"`

	// Budpay credentials
	BudpaySecretKey string `mapstructure:"BUDPAY_SECRET_KEY"`
	BudpayPublicKey string `mapstructure:"BUDPAY_PUB_KEY"`

	EnableEmail bool `mapstructure:"ENABLE_EMAIL"` // value comes from the database
	EnableSMS   bool `mapstructure:"ENABLE_SMS"`   // value comes from the database

	AfricaStalkingUsername string `mapstructure:"AFRICASTALKING_USERNAME"`
	AfricaStalkingAPIKey   string `mapstructure:"AFRICASTALKING_API_KEY"`
	AfricaStalkingAPIURL   string `mapstructure:"AFRICASTALKING_API_URL"`
	AfricaStalkingFromName string `mapstructure:"AFRICASTALKING_FROM"`

	PasspointBaseURL     string `mapstructure:"PASSPOINT_BASE_URL"`
	PasspointAuthBaseURL string `mapstructure:"PASSPOINT_AUTH_BASE_URL"`
	PasspointMerchantID  string `mapstructure:"PASSPOINT_MERCHANT_ID"`
	AdminEmail           string `mapstructure:"ADMIN_EMAIL"`

	// AutoProcessBankInflowLimit is the amount that will be used to determine if a bank inflow should be processed automatically
	AutoProcessBankInflowLimit int64 `mapstructure:"AUTO_PROCESS_BANK_INFLOW_LIMIT"`

	VeriffBaseURL      string `mapstructure:"VERIFF_BASE_URL"`
	VeriffClientKey    string `mapstructure:"VERIFF_CLIENT_KEY"`
	VeriffClientSecret string `mapstructure:"VERIFF_CLIENT_SECRET"`
	VeriffCallbackBase string `mapstructure:"VERIFF_CLIENT_CALLBACK_BASE"`

	ByBitP2PEndpoint         string `mapstructure:"BYBIT_P2P_ENDPOINT"`
	CurrencyLayerAccessKey   string `mapstructure:"CURRENCY_LAYER_ACCESS_KEY"`
	CurrencyLayerAPIEndpoint string `mapstructure:"CURRENCY_LAYER_API_ENDPOINT"`
	CurrencyLayerMocked      bool   `mapstructure:"CURRENCY_LAYER_IS_MOCKED"`

	EasyEuroMasterWalletID  string `mapstructure:"EASY_EURO_MASTER_WALLET_ID"`
	EasyEuroMasterAccountID string `mapstructure:"EASY_EURO_MASTER_ACCOUNT_ID"`
	EasyEuroAppKey          string `mapstructure:"EASY_EURO_APP_KEY"`
	EasyEuroAPPSecret       string `mapstructure:"EASY_EURO_APP_SECRET"`
	EasyEuroAPIURL          string `mapstructure:"EASY_EURO_API_URL"`

	// ZylaLabs
	ZylaLabsAPIKey     string `mapstructure:"ZYLALABS_API_KEY"`
	EasyEuroWebhookURL string `mapstructure:"EASY_EURO_WEBHOOK_URL"`

	// OTP DURATION
	OTPDuration time.Duration `mapstructure:"OTP_DURATION"`

	// Fidelity bank
	FidelityBankAPIKey    string `mapstructure:"FIDELITY_BANK_API_KEY"`
	FidelityBankAPISecret string `mapstructure:"FIDELITY_BANK_API_SECRET"`
}

func Load(p string) (cfg Config, err error) {
	return loader(p, ".env")
}

func LoadWithPath(p string, env string) (cfg Config, err error) {
	return loader(p, env)
}

func loader(p string, env string) (cfg Config, err error) {
	viper.AddConfigPath(p)
	viper.SetConfigName(env)
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	return
}
