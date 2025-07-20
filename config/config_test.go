package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

func TestSetLogging_debug(t *testing.T) {
	c := Config{}
	c.Log.Level = "debug"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.DebugLevel {
		t.Fatalf("Expected debug level, got %v", log.GetLevel())
	}
}

func TestLoadConfig(t *testing.T) {
	var tmpConfig configFlags
	parser := flags.NewParser(&tmpConfig, flags.Default)
	if _, err := parser.ParseArgs([]string{"--db-host=test", "--port=1234"}); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatalf("Failed to parse flags: %s", err)
	}
	compareConfig := configFlags{
		Log: LogConfig{
			Level:  "info",
			Format: "plain",
		},
		ConfigFilePath: "",
		DB: DBConfig{
			Host:         "test",
			Port:         5432,
			User:         "gorm",
			Password:     "",
			Name:         "gorm",
			SSLMode:      "require",
			NoSync:       false,
			SyncInterval: 1,
		},
		AWS: AWSConfig{
			AccessKey:       "",
			SecretAccessKey: "",
			DynamoDBTable:   "",
		},
		S3: S3BucketConfig{
			Bucket:         "",
			KeyPrefix:      "",
			FileExtension:  []string{".tfstate"},
			ForcePathStyle: false,
		},
		TFE: TFEConfig{
			Address:      "",
			Token:        "",
			Organization: "",
		},
		GCP: GCPConfig{
			GCSBuckets: nil,
			GCPSAKey:   "",
		},
		Gitlab: GitlabConfig{
			Address: "https://gitlab.com",
			Token:   "",
		},
		Kubernetes: KubernetesConfig{
			SecretSuffix:           "tfstate",
			Namespace:              "default",
			Labels:                 nil,
			ConfigPath:             "",
			ConfigContext:          "",
			ConfigContextAuthInfo:  "",
			ConfigContextCluster:   "",
			InClusterConfig:        false,
		},
		Web: WebConfig{
			Port:        1234,
			SwaggerPort: 8081,
			BaseURL:     "/",
			LogoutURL:   "",
		},
	}

	if !reflect.DeepEqual(tmpConfig, compareConfig) {
		t.Errorf(
			"TestLoadConfig() -> \n\ngot:\n%v,\n\nwant:\n%v",
			spew.Sdump(tmpConfig),
			spew.Sdump(compareConfig),
		)
	}
}

func TestLoadConfigFromYaml(t *testing.T) {
	var config Config
	os.Setenv("AWS_DEFAULT_REGION", "test-region")
	defer os.Unsetenv("AWS_DEFAULT_REGION")
	config.LoadConfigFromYaml("config_test.yml")
	compareConfig := Config{
		Log: LogConfig{
			Level:  "error",
			Format: "json",
		},
		ConfigFilePath: "config_test.yml",
		DB: DBConfig{
			Host:         "postgres",
			Port:         15432,
			User:         "terraboard-user",
			Password:     "terraboard-pass",
			Name:         "terraboard-db",
			SSLMode:      "require",
			NoSync:       true,
			SyncInterval: 1,
		},
		AWS: []AWSConfig{
			{
				AccessKey:       "root",
				SecretAccessKey: "mypassword",
				DynamoDBTable:   "terraboard-dynamodb",
				Region:          "test-region",
				S3: []S3BucketConfig{{
					Bucket:         "terraboard-bucket",
					KeyPrefix:      "test/",
					FileExtension:  []string{".tfstate"},
					ForcePathStyle: true,
				}},
			},
		},
		TFE: []TFEConfig{
			{
				Address:      "https://tfe.example.com",
				Token:        "foo",
				Organization: "bar",
			},
		},
		GCP: []GCPConfig{
			{
				GCSBuckets: []string{"my-bucket-1", "my-bucket-2"},
				GCPSAKey:   "/path/to/key",
			},
		},
		Gitlab: []GitlabConfig{
			{
				Address: "https://gitlab.example.com",
				Token:   "foo",
			},
		},
		Web: WebConfig{
			Port:        39090,
			SwaggerPort: 8081,
			BaseURL:     "/test/",
			LogoutURL:   "/test-logout",
		},
	}

	if !reflect.DeepEqual(config, compareConfig) {
		t.Errorf(
			"TestLoadConfig() -> \n\ngot:\n%v,\n\nwant:\n%v",
			spew.Sdump(config),
			spew.Sdump(compareConfig),
		)
	}
}

func TestSetLogging_info(t *testing.T) {
	c := Config{}
	c.Log.Level = "info"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.InfoLevel {
		t.Fatalf("Expected info level, got %v", log.GetLevel())
	}
}

func TestSetLogging_warn(t *testing.T) {
	c := Config{}
	c.Log.Level = "warn"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.WarnLevel {
		t.Fatalf("Expected warn level, got %v", log.GetLevel())
	}
}

func TestSetLogging_error(t *testing.T) {
	c := Config{}
	c.Log.Level = "error"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.ErrorLevel {
		t.Fatalf("Expected error level, got %v", log.GetLevel())
	}
}

func TestSetLogging_fatal(t *testing.T) {
	c := Config{}
	c.Log.Level = "fatal"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.FatalLevel {
		t.Fatalf("Expected fatal level, got %v", log.GetLevel())
	}
}

func TestSetLogging_panic(t *testing.T) {
	c := Config{}
	c.Log.Level = "panic"
	c.Log.Format = "plain"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if log.GetLevel() != log.PanicLevel {
		t.Fatalf("Expected panic level, got %v", log.GetLevel())
	}
}

func TestSetLogging_wronglevel(t *testing.T) {
	c := Config{}
	c.Log.Level = "wrong"
	c.Log.Format = "plain"
	err := c.SetupLogging()

	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	expectedError := "Wrong log level 'wrong'"

	if err.Error() != expectedError {
		t.Fatalf("Expected %s, got %s", expectedError, err.Error())
	}
}

func TestSetLogging_json(t *testing.T) {
	c := Config{}
	c.Log.Level = "debug"
	c.Log.Format = "json"
	err := c.SetupLogging()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestSetLogging_wrongformat(t *testing.T) {
	c := Config{}
	c.Log.Level = "debug"
	c.Log.Format = "yaml"
	err := c.SetupLogging()

	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}

	expectedError := "Wrong log format 'yaml'"

	if err.Error() != expectedError {
		t.Fatalf("Expected %s, got %s", expectedError, err.Error())
	}
}

func TestDBEnvVarOverride(t *testing.T) {
	// Create a temp config file
	content := []byte(`
database:
  host: db-host-from-yaml
  port: 1234
  user: user-from-yaml
  password: password-from-yaml
  name: db-name-from-yaml
  sslmode: "disable"
`)
	tmpfile, err := ioutil.TempFile("", "testconfig.*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set env vars to override
	os.Setenv("DB_HOST", "db-host-from-env")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "user-from-env")
	os.Setenv("DB_PASSWORD", "password-from-env")
	os.Setenv("DB_NAME", "db-name-from-env")
	os.Setenv("DB_SSLMODE", "require")

	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")
	defer os.Unsetenv("DB_USER")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("DB_NAME")
	defer os.Unsetenv("DB_SSLMODE")

	// Manipulate os.Args to point to the temp config file
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--config-file", tmpfile.Name()}

	config := LoadConfig("test-version")

	if config.DB.Host != "db-host-from-env" {
		t.Errorf("Expected DB.Host 'db-host-from-env', got '%s'", config.DB.Host)
	}
	if config.DB.Port != 5432 {
		t.Errorf("Expected DB.Port 5432, got '%d'", config.DB.Port)
	}
	if config.DB.User != "user-from-env" {
		t.Errorf("Expected DB.User 'user-from-env', got '%s'", config.DB.User)
	}
	if config.DB.Password != "password-from-env" {
		t.Errorf("Expected DB.Password 'password-from-env', got '%s'", config.DB.Password)
	}
	if config.DB.Name != "db-name-from-env" {
		t.Errorf("Expected DB.Name 'db-name-from-env', got '%s'", config.DB.Name)
	}
	if config.DB.SSLMode != "require" {
		t.Errorf("Expected DB.SSLMode 'require', got '%s'", config.DB.SSLMode)
	}
}
