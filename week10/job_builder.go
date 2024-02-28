package main

type CronJobFuncAdapter func() error

func (c CronJobFuncAdapter) Run() {
	_ = c()
}
