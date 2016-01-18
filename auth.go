package goapi;

type ValidateToken func(token string, log Logger) (userId int, userName string, err error)
