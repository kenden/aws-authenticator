// Package main allows to update the file ~/.aws/credentials with a valid AWS session token.
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
	getTokenProfile = "getToken"
	tempFilePath    = "~/.aws/credentials.tmp"
	origFilePath    = "~/.aws/credentials"
	defaultProfile  = "default"
)

// Config represent the AWS user's connection information
type Config struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSSessionToken    string
	AWSAssumeRoleArn   string
	AWSDefaultProfile  string
}

// NewConfig returns an empty Config
func NewConfig() *Config {
	return &Config{
		AWSAssumeRoleArn:  "", // TODO needed? It should be empty already
		AWSDefaultProfile: defaultProfile,
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
		Credentials: credentials.NewSharedCredentials("", getTokenProfile),
	})

	var iam *iam.User
	if iam, err = getCurrentUser(session); err != nil {
		log.Fatal(err.Error())
	}

	args := os.Args[1:]

	if conf, err = getSessionToken(session, iam, token, args...); err != nil {
		log.Fatal("Error while creating session: ", err.Error())
	}

	tempPath := expandPath(tempFilePath)
	origPath := expandPath(origFilePath)

	// TODO Do we really need this check? Isn't it done by expandPath() already ?
	if tempPath[:2] == "~/" {
		usr, _ := user.Current()
		tempPath = usr.HomeDir + tempPath[1:]
	}

	c, _ := openConfig(&origPath)
	writeTempConfig(c, conf, tempPath)

	swapFiles(tempPath, origPath)
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

	assumeRole := ""
	defaultProfile := defaultProfile

	if len(args) > 0 {
		assumeRole = args[0]
		if len(args) > 1 {
			defaultProfile = args[1]
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
		AWSAccessKeyID:     *out.Credentials.AccessKeyId,
		AWSAssumeRoleArn:   assumeRole,
		AWSDefaultProfile:  defaultProfile,
	}

	return config, err
}

// writeTempConfig save the config in parameter to file tmp
func writeTempConfig(c *config.Config, config *Config, tmp string) error {
	c.AddSection("default")
	c.AddOption("default", "aws_access_key_id", config.AWSAccessKeyID)
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
