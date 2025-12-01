package goocron

import "errors"

var (
	// ErrCronNotFound Cron 实例未找到
	ErrCronNotFound = errors.New("cron instance not found")
)

