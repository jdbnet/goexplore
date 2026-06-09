package keychain

import "github.com/zalando/go-keyring"

const serviceName = "goexplore"

func SetSecret(id string, secret string) error {
	key := "goexplore-" + id
	return keyring.Set(serviceName, key, secret)
}

func GetSecret(id string) (string, error) {
	key := "goexplore-" + id
	return keyring.Get(serviceName, key)
}

func DeleteSecret(id string) error {
	key := "goexplore-" + id
	return keyring.Delete(serviceName, key)
}
