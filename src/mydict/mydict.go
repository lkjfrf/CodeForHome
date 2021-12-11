package mydict

import "errors"

type Dictonary map[string]string

func (d Dictonary) Search(key string) (string, error) {
	result, ok := (d)[key]
	if ok {
		return result, nil
	}

	return "", errors.New("can't not found")
}

func (d Dictonary) Add(key string, value string) {
	_, ok := (d)[key]
	if !ok {
		(d)[key] = value
	}
}

func (d Dictonary) Delete(key string) {
	_, ok := (d)[key]
	if ok {
		delete(d, key)
	}
}
