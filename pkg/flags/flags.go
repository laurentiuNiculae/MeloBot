package flags

import "flag"

type Flags struct {
	CredentialFilePath *string
}

func GetFlags() Flags {
	var (
		credentialsFilePath string
	)

	flag.StringVar(&credentialsFilePath, "creds-file", "", "if specified, will choose the credentials from this file")

	flag.Parse()

	return Flags{
		CredentialFilePath: &credentialsFilePath,
	}
}
