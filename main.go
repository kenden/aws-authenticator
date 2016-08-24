package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/robfig/config"
)

const (
	GET_TOKEN_PROFILE = "getToken"
	TEMP_FILE_PATH    = "~/.aws/credentials.tmp"
	ORIG_FILE_PATH    = "~/.aws/credentials"
	DEFAULT_PROFILE   = "default"
)

type Config struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string
	AWSSessionToken    string
	AWSAssumeRoleArn   string
	AWSDefaultProfile  string
}

func NewConfig() *Config {
	return &Config{
		AWSAssumeRoleArn:  "",
		AWSDefaultProfile: DEFAULT_PROFILE,
	}
}

// main is the entry point of the program.
// It gets the MFA token from the user, get a token for the session and saves it
// to the credentials file.
func main() {

	token := getToken("Please enter MFA token: ")

	conf := NewConfig()
	var err error

	session := session.New(&aws.Config{
		Credentials: credentials.NewSharedCredentials("", GET_TOKEN_PROFILE),
	})

	var iam *iam.User
	if iam, err = getCurrentUser(session); err != nil {
		log.Fatal(err.Error())
	}

	args := os.Args[1:]

	if conf, err = getSessionToken(session, iam, token, args...); err != nil {
		log.Fatal("Error while creating session: ", err.Error())
	}

	temp_path := expandPath(TEMP_FILE_PATH)
	orig_path := expandPath(ORIG_FILE_PATH)

	// DO WE REALLY NEED THIS CHECK?
	if temp_path[:2] == "~/" {
		usr, _ := user.Current()
		temp_path = usr.HomeDir + temp_path[1:]
	}

	c, _ := openConfig(&orig_path)
	writeTempConfig(c, conf, temp_path)

	swapFiles(temp_path, orig_path)
	log.Println("Credentials has been updated!")
}

// expandPath replaces ~/' in the path given in parameter
// by the real path of the user's home directory
func expandPath(path string) string {
	if path[:2] == "~/" {
		usr, _ := user.Current()
		path = usr.HomeDir + path[1:]
	}

	return path
}

//openConfig reads the file in parameter and returns a Config struct
func openConfig(filename *string) (*config.Config, error) {
	return config.Read(*filename, "# ", "=", true, true)
}

// getToken prompts the user for the a token and returns it
func getToken(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf(prompt)
	scanner.Scan()

	return scanner.Text()
}

// getCurrentUser returns the IAM user.
func getCurrentUser(session *session.Session) (*iam.User, error) {
	svn := iam.New(session)
	user, err := svn.GetUser(&iam.GetUserInput{})

	return user.User, err
}

// getSessionToken gets a session token for the user. It returns a Config struct that includes it.
func getSessionToken(
	session *session.Session,
	user *iam.User,
	token string,
	args ...string) (*Config, error) {

	assume_role := ""
	default_profile := DEFAULT_PROFILE

	if len(args) > 0 {
		assume_role = args[0]
		if len(args) > 1 {
			default_profile = args[1]
		}
	}

	sn := strings.Replace(*user.Arn, ":user/", ":mfa/", 1)
	in := sts.GetSessionTokenInput{
		TokenCode:    &token,
		SerialNumber: &sn,
	}

	out, err := sts.New(session).GetSessionToken(&in)
	if err != nil {
		return nil, err
	}

	config := &Config{
		AWSSessionToken:    *out.Credentials.SessionToken,
		AWSSecretAccessKey: *out.Credentials.SecretAccessKey,
		AWSAccessKeyId:     *out.Credentials.AccessKeyId,
		AWSAssumeRoleArn:   assume_role,
		AWSDefaultProfile:  default_profile,
	}

	return config, err
}

// writeTempConfig save the config in parameter to file tmp
func writeTempConfig(c *config.Config, config *Config, tmp string) error {
	c.AddSection("default")
	c.AddOption("default", "aws_access_key_id", config.AWSAccessKeyId)
	c.AddOption("default", "aws_session_token", config.AWSSessionToken)
	c.AddOption("default", "aws_secret_access_key", config.AWSSecretAccessKey)
	c.AddOption("default", "role_arn", config.AWSAssumeRoleArn)
	c.AddOption("default", "source_profile", config.AWSDefaultProfile)

	// role_arn and source_profile only make sense if A
	if config.AWSAssumeRoleArn == "" {
		c.RemoveOption("default", "role_arn")
		c.RemoveOption("default", "source_profile")
	}

	return c.WriteFile(tmp, 0644, "Updated by Boynux authenticator")
}

// swapFiles renames file orig into file temp
func swapFiles(orig, temp string) error {
	return os.Rename(orig, temp)
}
