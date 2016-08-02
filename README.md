# aws-authenticator
MFA authenticator for AWS Cli

Basic MFA authenticator with assume role feature.

Usage:

  $ aws-authenticator

Then enter MFA token at the prompt.
  
Note: There should be a "getToken" profile in your `~/.aws/credentials` with at least two user policies:
  
* GetUser (for current user)
* Get user token with MFA

Like:

    {
        "Sid": "AllowUsersMFAForOwnCredentials",
        "Effect": "Allow",
        "Action": [
            "iam:GetSessionToken",
            "iam:GetUser"
        ],
        "Resource": [
            "arn:aws:iam::<AWS-ACCOUNT-ID>:mfa/${aws:username}"
        ]
    }
